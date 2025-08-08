package service

import (
	"context"
	"log/slog"
	"subscription/config"
	"subscription/internal/model"
	"subscription/internal/repository/postgres"
	"time"
)

// SubscriptionService описывает все кейсы работы подписок
type SubscriptionService interface {
	CreateSubscription(ctx context.Context, sub model.Subscription) (model.Subscription, error)

	GetSubscription(ctx context.Context, id int) (model.Subscription, error)

	ListSubscriptions(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error)

	UpdateSubscription(ctx context.Context, sub model.Subscription) error

	DeleteSubscription(ctx context.Context, id int) error

	Sum(ctx context.Context, userID, serviceName string, startPeriod, endPeriod time.Time) (int, error)

	Ping(ctx context.Context) error
}

type SubscriptionSvc struct {
	repo   SubscriptionService
	logger *slog.Logger
	config *config.Config
}

func NewSubscriptionService(repo SubscriptionService, logger *slog.Logger, config *config.Config) *SubscriptionSvc {
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

// проверяем имплиментацию
var _ SubscriptionService = (*postgres.Storage)(nil)
