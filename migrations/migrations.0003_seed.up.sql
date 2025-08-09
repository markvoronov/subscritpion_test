-- +migrate Up
INSERT INTO subscriptions (user_id, service_name, price, start_date)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'netflix', 10, DATE_TRUNC('month', NOW()) - INTERVAL '2 months'),
    ('22222222-2222-2222-2222-222222222222', 'spotify', 5, DATE_TRUNC('month', NOW()) - INTERVAL '1 month')
    ON CONFLICT DO NOTHING;
