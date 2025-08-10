package service

import (
	"context"
	"fmt"
	"log/slog"
	"subscription/internal/config"
	"subscription/internal/model"
	"subscription/internal/repository/postgres"
	"time"
)

// SubscriptionRepository — контракт ХРАНИЛИЩА данных
type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, sub model.Subscription) (model.Subscription, error)

	GetSubscription(ctx context.Context, id int) (model.Subscription, error)

	ListSubscriptions(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error)

	UpdateSubscription(ctx context.Context, sub model.Subscription) error

	DeleteSubscription(ctx context.Context, id int) error

	Sum(ctx context.Context, userID, serviceName string, startPeriod, endPeriod time.Time) (int, error)

	Ping(ctx context.Context) error
}

type SubscriptionSvc struct {
	repo   SubscriptionRepository
	logger *slog.Logger
	config *config.Config
}

func NewSubscriptionService(repo SubscriptionRepository, logger *slog.Logger, config *config.Config) *SubscriptionSvc {
	return &SubscriptionSvc{
		repo:   repo,
		logger: logger,
		config: config,
	}
}

func (s *SubscriptionSvc) CreateSubscription(ctx context.Context, sub model.Subscription) (model.Subscription, error) {
	// по хорошему на этом этапе нужно проверять, а не дубль ли это? Не пересекается ли подпсика по времени с новой?

	const op = "internal.service.CreateSubscription"
	log := s.logger.With(slog.String("op", op))

	if sub.Price < 0 {
		err := fmt.Errorf("price cannot be negative")
		log.Error("Can`t create new subscription", slog.String("error", err.Error()))
		return model.Subscription{}, err
	}

	sub, err := s.repo.CreateSubscription(ctx, sub)
	if err != nil {
		log.Error("Can`t create new subscription", slog.String("error", err.Error()))
		return sub, err
	}

	return sub, nil
}

func (s *SubscriptionSvc) GetSubscription(ctx context.Context, id int) (model.Subscription, error) {
	return s.repo.GetSubscription(ctx, id)
}

func (s *SubscriptionSvc) UpdateSubscription(ctx context.Context, sub model.Subscription) error {
	// по хорошему на этом этапе нужно проверять, чтобы подписка не пересекалась с другой от этого же пользователя
	// и сервиса

	return s.repo.UpdateSubscription(ctx, sub)
}

func (s *SubscriptionSvc) DeleteSubscription(ctx context.Context, id int) error {
	return s.repo.DeleteSubscription(ctx, id)
}

func (s *SubscriptionSvc) Sum(ctx context.Context, userID, serviceName string, startPeriod, endPeriod time.Time) (int, error) {
	return s.repo.Sum(ctx, userID, serviceName, startPeriod, endPeriod)
}

func (s *SubscriptionSvc) ListSubscriptions(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error) {
	return s.repo.ListSubscriptions(ctx, userID, serviceName)
}

func (s *SubscriptionSvc) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

// проверяем имплиментацию
var _ SubscriptionRepository = (*postgres.Storage)(nil)
