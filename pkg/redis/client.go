package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/linkyfish/kxl_backend_go/internal/config"
)

func NewClient(cfg *config.Config) (*redis.Client, error) {
	var opt *redis.Options
	if cfg.Redis.URL != "" {
		parsed, err := redis.ParseURL(cfg.Redis.URL)
		if err != nil {
			return nil, err
		}
		opt = parsed
	} else {
		opt = &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}
	}

	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

