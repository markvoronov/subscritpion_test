-- +migrate Down
DROP INDEX IF EXISTS idx_subscriptions_user;
DROP INDEX IF EXISTS idx_subscriptions_service;
DROP INDEX IF EXISTS idx_subscriptions_period;
