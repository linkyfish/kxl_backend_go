package middleware

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	kxlcfg "github.com/linkyfish/kxl_backend_go/internal/config"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/labstack/echo/v4"
)

type RedisRateLimiter struct {
	Client       *redis.Client
	KeyPrefix    string
	Window       time.Duration
	MaxRequests  int64
}

var incrExpireScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if tonumber(current) == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return current
`)

func (rl *RedisRateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl == nil || rl.Client == nil || rl.Window <= 0 || rl.MaxRequests <= 0 {
				return next(c)
			}

			ip := c.RealIP()
			key := fmt.Sprintf("%s:%s", rl.KeyPrefix, ip)

			ctx := c.Request().Context()
			if ctx == nil {
				ctx = context.Background()
			}

			n, err := incrExpireScript.Run(ctx, rl.Client, []string{key}, int64(rl.Window.Seconds())).Int64()
			if err == nil && n > rl.MaxRequests {
				return kxlerrors.TooManyRequests()
			}

			// Fail-open on Redis errors to avoid taking down the API.
			return next(c)
		}
	}
}

// RateLimit implements the same endpoint-aware policy as the PHP backend:
// - Login endpoints: key by ip + identifier fragment.
// - Upload endpoints: key by user/admin id (if available), otherwise by ip.
func RateLimit(client *redis.Client, cfg *kxlcfg.Config) echo.MiddlewareFunc {
	windowLogin := 60
	maxLogin := int64(20)
	windowUpload := 60
	maxUpload := int64(30)
	if cfg != nil {
		if cfg.Security.RateLimitLoginWindowSeconds > 0 {
			windowLogin = cfg.Security.RateLimitLoginWindowSeconds
		}
		if cfg.Security.RateLimitLoginMaxAttempts > 0 {
			maxLogin = int64(cfg.Security.RateLimitLoginMaxAttempts)
		}
		if cfg.Security.RateLimitUploadWindowSeconds > 0 {
			windowUpload = cfg.Security.RateLimitUploadWindowSeconds
		}
		if cfg.Security.RateLimitUploadMaxRequests > 0 {
			maxUpload = int64(cfg.Security.RateLimitUploadMaxRequests)
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if client == nil {
				return next(c)
			}

			req := c.Request()
			if req == nil {
				return next(c)
			}
			method := strings.ToUpper(req.Method)
			path := req.URL.Path

			if method == http.MethodPost && (path == "/api/v1/auth/login" || path == "/api/admin/auth/login") {
				ip := c.RealIP()
				identifier := extractIdentifier(c)
				actor := "user"
				if strings.HasPrefix(path, "/api/admin/") {
					actor = "admin"
				}
				key := fmt.Sprintf("rl:login:%s:%s:%s", actor, ip, redisKeyFragment(identifier))
				if err := enforceRateLimit(req.Context(), client, key, int64(windowLogin), maxLogin); err != nil {
					return err
				}
			}

			if method == http.MethodPost && (path == "/api/upload/image" || path == "/api/upload/video" || path == "/api/admin/upload/image" || path == "/api/admin/upload/video") {
				kind := "image"
				if strings.HasSuffix(path, "/video") {
					kind = "video"
				}
				actor := "user"
				if strings.HasPrefix(path, "/api/admin/") {
					actor = "admin"
				}

				idKey := "current_" + actor + "_id"
				id, _ := c.Get(idKey).(string)
				var key string
				if id != "" {
					key = fmt.Sprintf("rl:upload:%s:%s:%s", actor, id, kind)
				} else {
					ip := c.RealIP()
					key = fmt.Sprintf("rl:upload:%s:%s:%s", actor, ip, kind)
				}

				if err := enforceRateLimit(req.Context(), client, key, int64(windowUpload), maxUpload); err != nil {
					return err
				}
			}

			return next(c)
		}
	}
}

func enforceRateLimit(ctx context.Context, client *redis.Client, key string, windowSeconds int64, maxRequests int64) error {
	if windowSeconds <= 0 || maxRequests <= 0 {
		return nil
	}
	n, err := incrExpireScript.Run(ctx, client, []string{key}, windowSeconds).Int64()
	if err != nil {
		// Fail-open on Redis errors.
		return nil
	}
	if n > maxRequests {
		return kxlerrors.New(kxlerrors.CodeTooManyRequests, "too many requests: rate limit exceeded", http.StatusTooManyRequests, nil)
	}
	return nil
}

func redisKeyFragment(input string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(input))
	b64 = strings.ReplaceAll(b64, "+", "-")
	b64 = strings.ReplaceAll(b64, "/", "_")
	b64 = strings.ReplaceAll(b64, "=", "")
	return b64
}

func extractIdentifier(c echo.Context) string {
	// Prefer form value.
	if v := c.FormValue("identifier"); v != "" {
		return v
	}
	// Try JSON body (non-destructive read).
	req := c.Request()
	if req == nil {
		return ""
	}
	ct := req.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		return ""
	}
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return ""
	}
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var payload map[string]interface{}
	if json.Unmarshal(bodyBytes, &payload) != nil {
		return ""
	}
	if v, ok := payload["identifier"].(string); ok {
		return v
	}
	return ""
}
