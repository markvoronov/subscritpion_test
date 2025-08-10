package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"subscription/internal/config"
	"subscription/internal/handler"
	mwLogger "subscription/internal/middleware/logger"
	"subscription/internal/repository/postgres"
	"subscription/internal/service"
	"subscription/migrations"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.LoadConfig()
	logger := setupLogger(cfg.Env)
	logger.Debug("debug messages are enable")

	// 1) миграции до подключения пула приложения
	if err := migrations.RunMigrations(cfg, logger); err != nil {
		logger.Error("migrations failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 2) repo
	repo, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		logger.Error("db connect failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			logger.Error("failed to close repo", slog.String("error", err.Error()))
		}
	}()

	// 3) services
	services := service.NewSubscriptionService(repo, logger, cfg)
	h := handler.NewHandler(services, logger)

	// 4) router + middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(mwLogger.New(logger))
	r.Use(middleware.Timeout(cfg.HTTPServer.Timeout))

	// healthz
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 5) Swagger
	r.Get("/swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/openapi.yaml")
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/openapi.yaml"),
	))

	// 6) API
	r.Post("/subscriptions", h.CreateSubscription)
	r.Get("/subscriptions/{id}", h.GetSubscription)
	r.Get("/subscriptions", h.ListSubscriptions)
	r.Put("/subscriptions/{id}", h.UpdateSubscription)
	r.Delete("/subscriptions/{id}", h.DeleteSubscription)
	r.Get("/subscriptions/summary", h.SumSubscriptions)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "resource not found", http.StatusNotFound)
		logger.Info("not found", "path", r.URL.Path)
	})

	// 7) http.Server c таймаутами + graceful shutdown
	addr := net.JoinHostPort(cfg.HTTPServer.Address, cfg.HTTPServer.HTTPPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	logger.Info("service starting", "addr", addr)

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", slog.String("error", err.Error()))
		}
	}()

	// Ожидаем сигнал и красиво гасим сервер
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPServer.ShutdownTimeout)

	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		logger.Error("server shutdown error", slog.String("error", err.Error()))
	}
	logger.Info("server stopped")

}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
