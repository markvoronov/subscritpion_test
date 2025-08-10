-- +migrate Down
DELETE FROM subscriptions
WHERE user_id IN (
                  '11111111-1111-1111-1111-111111111111',
                  '22222222-2222-2222-2222-222222222222'
    )
  AND service_name IN ('netflix','spotify');
