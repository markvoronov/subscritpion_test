# subscription
Тестовое задание от Effective Mobile

REST-сервис для агрегации данных об онлайн-подписках пользователей. Проект выполнен по ТЗ «Junior Golang Developer» (Effective Mobile).

## Возможности
CRUDL для подписок:
- `service_name` — название сервиса
- `price` — стоимость/мес (₽)
- `user_id` — UUID пользователя
- `start_date` — месяц/год начала (ввод в формате `MM-YYYY`)
- `end_date` — опционально месяц/год окончания (ввод в формате `MM-YYYY`)

Дополнительно:
- Подсчёт суммарной стоимости подписок за период с фильтрами по `user_id` и `service_name`
- PostgreSQL + миграции
- Логирование (`slog`) и middleware
- Конфиг через YAML
- Swagger-документация
- Запуск через Docker Compose

## Стек
Go, Chi, PostgreSQL, golang-migrate, slog, http-swagger, Docker/Compose.

## Быстрый старт
```bash
git clone <repo-url>
cd subscription
docker-compose up --build
```

## Доступ к сервису
Сервис: http://localhost:8080

Swagger UI: http://localhost:8080/swagger/index.html

Спецификация: http://localhost:8080/swagger/openapi.yaml

## Конфигурация
По умолчанию читается ./config/config.yaml.
Путь можно переопределить переменной CONFIG_PATH.

## Логи
Используется slog с уровнями, формат зависит от ENV:

local — текстовый, DEBUG

dev — JSON, DEBUG

prod — JSON, INFO

## Заметки
Параметр CONFIG_PATH позволяет указать путь к конфигу при запуске.

Валидация дат: входящие start_date / end_date ожидаются строго как MM-YYYY; парсятся через time.Parse("01-2006").
