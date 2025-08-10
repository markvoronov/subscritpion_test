-- +migrate Up
CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service ON subscriptions (service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_period ON subscriptions (start_date, end_date);
