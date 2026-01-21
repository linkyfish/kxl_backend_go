package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kxlcfg "github.com/linkyfish/kxl_backend_go/internal/config"
	"github.com/linkyfish/kxl_backend_go/internal/router"
	"github.com/linkyfish/kxl_backend_go/pkg/db"
	kxlredis "github.com/linkyfish/kxl_backend_go/pkg/redis"
	"github.com/linkyfish/kxl_backend_go/pkg/session"
)

func main() {
	cfg, err := kxlcfg.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	gormDB, err := db.ConnectPostgres(cfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	redisClient, err := kxlredis.NewClient(cfg)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer func() { _ = redisClient.Close() }()

	sess := session.NewManager(redisClient, cfg)

	e := router.New(router.Deps{
		Cfg:   cfg,
		DB:    gormDB,
		Redis: redisClient,
		Sess:  sess,
	})

	// Graceful shutdown.
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("echo start: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

