package postgres

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"strings"
	"subscription/config"
	"subscription/internal/model"
	"time"
)

type Storage struct {
	db *sql.DB
}

func NewPostgresDB(cfg *config.Config) (*Storage, error) {

	ps := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode)

	db, err := sql.Open("pgx", ps)

	if err != nil {
		return nil, err
	}

	s := &Storage{db: db}

	if err := s.Ping(context.Background()); err != nil {
		db.Close()      // НЕ забываем закрыть открытое соединение
		return nil, err // и возвращаем ошибку дальше
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

func (r *Storage) CreateSubscription(ctx context.Context, s model.Subscription) (model.Subscription, error) {
	query := `
        INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `
	return s, r.db.QueryRowContext(ctx, query, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).
		Scan(&s.ID)
}
func (r *Storage) GetSubscription(ctx context.Context, id int) (model.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date
        FROM subscriptions
        WHERE id = $1
    `
	var s model.Subscription
	var endDate sql.NullTime
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &endDate)
	if err != nil {
		return s, err
	}
	if endDate.Valid {
		s.EndDate = &endDate.Time
	}
	return s, nil
}

// ListSubscriptions Архитектурно правильно было бы делать пагинацию
func (r *Storage) ListSubscriptions(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error) {
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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*model.Subscription
	for rows.Next() {
		var s model.Subscription
		var endDate sql.NullTime
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &endDate); err != nil {
			return nil, err
		}
		if endDate.Valid {
			s.EndDate = &endDate.Time
		}
		subs = append(subs, &s)
	}
	return subs, nil
}

func (r *Storage) UpdateSubscription(ctx context.Context, s model.Subscription) error {
	query := `
        UPDATE subscriptions
        SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
        WHERE id = $6
    `
	_, err := r.db.ExecContext(ctx, query, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate, s.ID)
	return err
}

func (r *Storage) DeleteSubscription(ctx context.Context, id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *Storage) Sum(ctx context.Context, userID, serviceName string, startPeriod, endPeriod time.Time) (int, error) {
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
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&sum)
	return sum, err
}
