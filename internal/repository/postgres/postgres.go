package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"strings"
	"subscription/internal/config"
	"subscription/internal/model"
	"time"
)

type Storage struct {
	db *sql.DB
}

func NewPostgresDB(cfg *config.Config) (*Storage, error) {

	ps := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode)

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Устанавливаем таймауты
	// ConnLifetime задаётся в секундах в конфиге
	db.SetConnMaxLifetime(cfg.Database.Pool.ConnLifetime)
	db.SetMaxOpenConns(cfg.Database.Pool.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.Pool.MaxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s := &Storage{db: db}
	if err := s.Ping(ctx); err != nil {
		// Не забываем закрыть открытое соединение
		if cerr := db.Close(); cerr != nil {
			return nil, errors.Join(err, cerr)
		}

		return nil, err
	}

	return s, nil

}

func (s *Storage) Ping(ctx context.Context) error {

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Используем PingContext вместо Ping
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

func (s *Storage) CreateSubscription(ctx context.Context, sub model.Subscription) (model.Subscription, error) {
	query := `
        INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `
	return sub, s.db.QueryRowContext(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		Scan(&sub.ID)
}

func (s *Storage) GetSubscription(ctx context.Context, id int) (model.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date
        FROM subscriptions
        WHERE id = $1
    `
	var sub model.Subscription
	var endDate sql.NullTime
	row := s.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &endDate)
	if err != nil {
		return sub, err
	}
	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}
	return sub, nil
}

// ListSubscriptions Архитектурно правильно было бы делать пагинацию
func (s *Storage) ListSubscriptions(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error) {
	var where []string
	var args []interface{}
	idx := 1

	if userID != "" {
		where = append(where, fmt.Sprintf("user_id = $%d", idx))
		args = append(args, userID)
		idx++
	}
	if serviceName != "" {
		where = append(where, fmt.Sprintf("service_name = $%d", idx))
		args = append(args, serviceName)
		idx++
	}

	query := `
        SELECT id, service_name, price, user_id, start_date, end_date
        FROM subscriptions
    `
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY id"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Будем аккумулировать ошибку закрытия в именованном ретёрне.
	var retErr error
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			retErr = errors.Join(retErr, fmt.Errorf("rows.Close: %w", cerr))
		}
	}()

	var subs []*model.Subscription
	for rows.Next() {
		var s model.Subscription
		var endDate sql.NullTime

		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &endDate); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		if endDate.Valid {
			s.EndDate = &endDate.Time
		}
		subs = append(subs, &s)
	}

	// Обязательная проверка ошибок итератора (в т.ч. контекстных).
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	// Если была только ошибка Close — вернём её отдельно.
	return subs, retErr
}

func (s *Storage) UpdateSubscription(ctx context.Context, sub model.Subscription) error {
	query := `
        UPDATE subscriptions
        SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
        WHERE id = $6
    `
	_, err := s.db.ExecContext(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID)
	return err
}

func (s *Storage) DeleteSubscription(ctx context.Context, id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *Storage) Sum(ctx context.Context, userID, serviceName string, startPeriod, endPeriod time.Time) (int, error) {
	var where []string
	var args []interface{}
	idx := 1

	// Фильтр по user_id
	if userID != "" {
		where = append(where, fmt.Sprintf("user_id = $%d", idx))
		args = append(args, userID)
		idx++
	}
	// Фильтр по service_name
	if serviceName != "" {
		where = append(where, fmt.Sprintf("service_name = $%d", idx))
		args = append(args, serviceName)
		idx++
	}
	// Подписка считается активной, если ее интервал пересекает выбранный период
	// Учитываем только подписки, у которых start_date <= endPeriod и (end_date IS NULL OR end_date >= startPeriod)
	where = append(where, fmt.Sprintf("start_date <= $%d", idx))
	args = append(args, endPeriod)
	idx++
	where = append(where, fmt.Sprintf("(end_date IS NULL OR end_date >= $%d)", idx))
	args = append(args, startPeriod)
	idx++

	// Если подписок не будет, вернем просто 0
	query := "SELECT COALESCE(SUM(price), 0) FROM subscriptions"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	var sum int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&sum)
	return sum, err
}

func (s *Storage) Close() error {
	return s.db.Close()
}
