package main

import (
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"log/slog"
	"net/http"
	"os"
	"subscription/config"
	"subscription/internal/handler"
	"subscription/internal/repository/postgres"
	"subscription/internal/service"
	migrations "subscription/migrations"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {

	cfg := config.LoadConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	logger.Debug("debug massages are enable")

	storage := initRepository(cfg, logger)
	runMigrations(cfg, logger)

	services := service.NewSubscriptionService(storage, logger, cfg)
	h := handler.NewHandler(services, logger)

	r := chi.NewRouter()
	r.Use(middleware.Timeout(2 * time.Second))

	// Swagger
	r.Get("/swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/openapi.yaml")
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/openapi.yaml"),
	))

	r.Post("/subscriptions", h.CreateSubscription)
	r.Get("/subscriptions/{id}", h.GetSubscription)
	r.Get("/subscriptions", h.ListSubscriptions)
	r.Put("/subscriptions/{id}", h.UpdateSubscription)
	r.Delete("/subscriptions/{id}", h.DeleteSubscription)
	r.Get("/subscriptions/summary", h.SumSubscriptions)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("subscriptions resource not found"))
		logger.Info("subscriptions resource not found", "path", r.URL)
	})
	logger.Info("service starting", "port", cfg.HTTPPort)
	http.ListenAndServe(":"+cfg.HTTPPort, r)
}

func initRepository(cfg *config.Config, Logger *slog.Logger) service.SubscriptionService {

	repo, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		Logger.Error("db connect failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	return repo

}

func runMigrations(cfg *config.Config, log *slog.Logger) {

	if err := migrations.RunMigrations(cfg, log); err != nil {
		log.Error("migrations failed", slog.String("err", err.Error()))
		os.Exit(1)
	}

}
