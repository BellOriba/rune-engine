package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BellOriba/rune-engine/internal/cache"
	"github.com/BellOriba/rune-engine/internal/handlers"
	"github.com/BellOriba/rune-engine/internal/logger"
	"github.com/BellOriba/rune-engine/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	log := logger.New()
	slog.SetDefault(log)

	log.Info("initializing RuneEngine backend",
		"env", os.Getenv("APP_ENV"),
		"log_level", os.Getenv("LOG_LEVEL"),
	)

	pool := worker.NewPool(5, 100, log)

	poolCtx, cancelPool := context.WithCancel(context.Background())
	pool.Start(poolCtx)

	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Cache, Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Cache, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Use(gin.Recovery())
	r.Use(logger.Middleware(log))

	setupRoutes(r, pool, log)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen and serve failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server gracefully...")

	cancelPool()
	pool.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
	}

	log.Info("server exited cleanly")
}

func setupRoutes(r *gin.Engine, pool *worker.Pool, log *slog.Logger) {
	rdb, err := cache.NewCache()
	if err != nil {
		log.Warn("Redis não disponível, operando sem cache", "error", err)
	}

	h := &handlers.ASCIIHandler{
		Pool: pool,
		Cache: rdb,
	}

	v1 := r.Group("/v1")
	{
		v1.GET("/health", handleHealth)
		v1.POST("/convert", h.ConvertImage)
		v1.POST("/stream", h.StreamGIF)
	}
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "up",
		"timestamp": time.Now().Unix(),
	})
}
