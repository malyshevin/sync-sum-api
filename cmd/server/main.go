package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/malyshevin/sync-sum-api/internal/config"
	"github.com/malyshevin/sync-sum-api/internal/httpapi"
	"github.com/malyshevin/sync-sum-api/internal/repository"
	"github.com/malyshevin/sync-sum-api/internal/service"
)

func main() {
	// Load env files (optional) before reading config
	_ = godotenv.Overload(".env")
	_ = godotenv.Overload(".env.local")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		logger.Error("failed to load config", slog.Any("err", err))
		os.Exit(1)
	}

	db, err := repository.OpenPostgres(context.Background(), cfg.Database.DSN(), logger)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("err", err))
		os.Exit(1)
	}
	defer db.Close()

	counterRepo := repository.NewCounterStore(db, logger)
	counterSvc := service.NewCounterService(counterRepo)

	r := chi.NewRouter()
	r.Use(httpapi.RequestLogger(logger))
	handler := httpapi.NewCounterHandler(counterSvc, logger)
	httpapi.RegisterRoutes(r, handler)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("server starting", slog.Int("port", cfg.HTTP.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", slog.Any("err", err))
	}
}
