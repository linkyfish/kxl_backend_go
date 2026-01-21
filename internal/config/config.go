package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Session  SessionConfig  `mapstructure:"session"`
	Security SecurityConfig `mapstructure:"security"`
	Uploads  UploadsConfig  `mapstructure:"uploads"`
	Cors     CorsConfig     `mapstructure:"cors"`
}

type AppConfig struct {
	Env   string `mapstructure:"env"`
	Debug bool   `mapstructure:"debug"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	URL            string `mapstructure:"url"`
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Name           string `mapstructure:"name"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	MaxConnections int    `mapstructure:"max_connections"`
}

type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type SessionConfig struct {
	UserCookieName  string `mapstructure:"user_cookie_name"`
	AdminCookieName string `mapstructure:"admin_cookie_name"`
	CookieSecure    bool   `mapstructure:"cookie_secure"`
	Prefix          string `mapstructure:"prefix"`
	UserTTLSeconds  int    `mapstructure:"user_ttl_seconds"`
	AdminTTLSeconds int    `mapstructure:"admin_ttl_seconds"`
}

type SecurityConfig struct {
	RateLimitLoginWindowSeconds  int `mapstructure:"rate_limit_login_window_seconds"`
	RateLimitLoginMaxAttempts    int `mapstructure:"rate_limit_login_max_attempts"`
	RateLimitUploadWindowSeconds int `mapstructure:"rate_limit_upload_window_seconds"`
	RateLimitUploadMaxRequests   int `mapstructure:"rate_limit_upload_max_requests"`
	RbacCacheTTLSeconds          int `mapstructure:"rbac_cache_ttl_seconds"`
}

type UploadsConfig struct {
	Dir           string `mapstructure:"dir"`
	ImageMaxBytes int64  `mapstructure:"image_max_bytes"`
	VideoMaxBytes int64  `mapstructure:"video_max_bytes"`
}

type CorsConfig struct {
	AllowOrigin string `mapstructure:"allow_origin"`
}

func Load() (*Config, error) {
	// Optional .env for local dev. It is fine if it doesn't exist.
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")
	v.AddConfigPath(".")

	// Defaults align with .env.example to keep behavior predictable.
	v.SetDefault("app.env", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8787)
	v.SetDefault("database.max_connections", 10)
	v.SetDefault("redis.db", 0)
	v.SetDefault("session.user_cookie_name", "kxl_user_session")
	v.SetDefault("session.admin_cookie_name", "kxl_admin_session")
	v.SetDefault("session.cookie_secure", false)
	v.SetDefault("session.prefix", "kxl_session:")
	v.SetDefault("session.user_ttl_seconds", 604800)
	v.SetDefault("session.admin_ttl_seconds", 7200)
	v.SetDefault("security.rate_limit_login_window_seconds", 60)
	v.SetDefault("security.rate_limit_login_max_attempts", 20)
	v.SetDefault("security.rate_limit_upload_window_seconds", 60)
	v.SetDefault("security.rate_limit_upload_max_requests", 30)
	v.SetDefault("security.rbac_cache_ttl_seconds", 300)
	v.SetDefault("uploads.dir", "uploads")
	v.SetDefault("uploads.image_max_bytes", int64(10*1024*1024))
	v.SetDefault("uploads.video_max_bytes", int64(500*1024*1024))
	v.SetDefault("cors.allow_origin", "*")

	// Read config file if present (config/config.yaml is recommended).
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	applyEnvOverrides(&cfg)

	if cfg.Server.Port <= 0 {
		return nil, fmt.Errorf("invalid SERVER_PORT: %d", cfg.Server.Port)
	}
	if cfg.Session.Prefix == "" {
		cfg.Session.Prefix = "kxl_session:"
	}

	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	// Server
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := getenvInt("SERVER_PORT"); v != nil {
		cfg.Server.Port = *v
	}

	// Database
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.Database.URL = v
		applyDatabaseURL(cfg, v)
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := getenvInt("DB_PORT"); v != nil {
		cfg.Database.Port = *v
	}
	if v := os.Getenv("DB_DATABASE"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := getenvInt("DB_MAX_CONNECTIONS"); v != nil {
		cfg.Database.MaxConnections = *v
	}

	// Redis
	if v := os.Getenv("REDIS_URL"); v != "" {
		cfg.Redis.URL = v
		applyRedisURL(cfg, v)
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := getenvInt("REDIS_PORT"); v != nil {
		cfg.Redis.Port = *v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := getenvInt("REDIS_DB"); v != nil {
		cfg.Redis.DB = *v
	}

	// Session
	if v := os.Getenv("SESSION_USER_COOKIE_NAME"); v != "" {
		cfg.Session.UserCookieName = v
	}
	if v := os.Getenv("SESSION_ADMIN_COOKIE_NAME"); v != "" {
		cfg.Session.AdminCookieName = v
	}
	if v := getenvBool("SESSION_COOKIE_SECURE"); v != nil {
		cfg.Session.CookieSecure = *v
	}
	if v := os.Getenv("SESSION_PREFIX"); v != "" {
		cfg.Session.Prefix = v
	}
	if v := getenvInt("SESSION_USER_TTL_SECONDS"); v != nil {
		cfg.Session.UserTTLSeconds = *v
	}
	if v := getenvInt("SESSION_ADMIN_TTL_SECONDS"); v != nil {
		cfg.Session.AdminTTLSeconds = *v
	}

	// Security
	if v := getenvInt("RATE_LIMIT_LOGIN_WINDOW_SECONDS"); v != nil {
		cfg.Security.RateLimitLoginWindowSeconds = *v
	}
	if v := getenvInt("RATE_LIMIT_LOGIN_MAX_ATTEMPTS"); v != nil {
		cfg.Security.RateLimitLoginMaxAttempts = *v
	}
	if v := getenvInt("RATE_LIMIT_UPLOAD_WINDOW_SECONDS"); v != nil {
		cfg.Security.RateLimitUploadWindowSeconds = *v
	}
	if v := getenvInt("RATE_LIMIT_UPLOAD_MAX_REQUESTS"); v != nil {
		cfg.Security.RateLimitUploadMaxRequests = *v
	}
	if v := getenvInt("RBAC_CACHE_TTL_SECONDS"); v != nil {
		cfg.Security.RbacCacheTTLSeconds = *v
	}

	// Uploads
	if v := os.Getenv("UPLOADS_DIR"); v != "" {
		cfg.Uploads.Dir = v
	}
	if v := getenvInt64("UPLOAD_IMAGE_MAX_BYTES"); v != nil {
		cfg.Uploads.ImageMaxBytes = *v
	}
	if v := getenvInt64("UPLOAD_VIDEO_MAX_BYTES"); v != nil {
		cfg.Uploads.VideoMaxBytes = *v
	}

	// CORS
	if v := os.Getenv("CORS_ALLOW_ORIGIN"); v != "" {
		cfg.Cors.AllowOrigin = v
	}
}

func applyDatabaseURL(cfg *Config, raw string) {
	u, err := url.Parse(raw)
	if err != nil {
		return
	}
	if u.User != nil {
		cfg.Database.Username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			cfg.Database.Password = p
		}
	}
	host := u.Hostname()
	if host != "" {
		cfg.Database.Host = host
	}
	if port := u.Port(); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	db := strings.TrimPrefix(u.Path, "/")
	if db != "" {
		cfg.Database.Name = db
	}
}

func applyRedisURL(cfg *Config, raw string) {
	u, err := url.Parse(raw)
	if err != nil {
		return
	}
	host := u.Hostname()
	if host != "" {
		cfg.Redis.Host = host
	}
	if port := u.Port(); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Redis.Port = p
		}
	}
	if u.User != nil {
		if p, ok := u.User.Password(); ok {
			cfg.Redis.Password = p
		}
	}
	if dbStr := strings.TrimPrefix(u.Path, "/"); dbStr != "" {
		if n, err := strconv.Atoi(dbStr); err == nil {
			cfg.Redis.DB = n
		}
	}
}

func getenvInt(key string) *int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &n
}

func getenvInt64(key string) *int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

func getenvBool(key string) *bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	v := strings.EqualFold(raw, "true") || raw == "1" || strings.EqualFold(raw, "yes")
	return &v
}

