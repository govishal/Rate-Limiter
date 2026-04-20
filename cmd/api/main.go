package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"rate-limiter/internal/config"
	"rate-limiter/internal/handler"
	"rate-limiter/internal/ratelimiter"
)

func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg := config.LoadConfig()
	log.Printf("config: addr=%s max_requests=%d window=%s",
		cfg.Addr, cfg.MaxRequests, cfg.Window)

	limiter := ratelimiter.New(cfg.MaxRequests, cfg.Window)

	engine := gin.Default()
	handler.NewHandler(limiter).RegisterRoutes(engine)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("Rate-limited API listening on %s", cfg.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	waitForShutdown(server)
}

const gracefulShutdownTimeout = 10 * time.Second

func waitForShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		return
	}
	log.Println("server stopped gracefully")
}
