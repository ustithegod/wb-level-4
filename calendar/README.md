# Calendar Service

HTTP-сервис небольшого календаря событий на `Go`, `chi`, `slog` и `in-memory` хранилище.

## Возможности

- `POST /create_event`
- `POST /update_event`
- `POST /delete_event`
- `GET /events_for_day`
- `GET /events_for_week`
- `GET /events_for_month`
- асинхронный логгер через канал
- воркер напоминаний через канал
- фоновая архивация старых событий

## Запуск

```bash
go mod tidy
go run ./cmd/calendar -port 8080
```

Переменные окружения:

- `PORT` — порт HTTP-сервера, по умолчанию `8080`
- `ARCHIVE_INTERVAL_MINUTES` — интервал архивации, по умолчанию `1`
- `LOG_BUFFER` — размер буфера асинхронного логгера, по умолчанию `128`

## Примеры запросов

Создание события:

```bash
curl -X POST http://localhost:8080/create_event \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": 1,
    "date": "2026-04-21",
    "event": "doctor appointment",
    "remind_at": "2026-04-21T09:30:00Z"
  }'
```

Обновление события:

```bash
curl -X POST http://localhost:8080/update_event \
  -H 'Content-Type: application/json' \
  -d '{
    "id": 1,
    "user_id": 1,
    "date": "2026-04-22",
    "event": "doctor appointment moved",
    "remind_at": "2026-04-22T08:00:00Z"
  }'
```

Удаление события:

```bash
curl -X POST http://localhost:8080/delete_event \
  -H 'Content-Type: application/json' \
  -d '{"id": 1}'
```

События на день:

```bash
curl 'http://localhost:8080/events_for_day?user_id=1&date=2026-04-21'
```

События на неделю:

```bash
curl 'http://localhost:8080/events_for_week?user_id=1&date=2026-04-21'
```

События на месяц:

```bash
curl 'http://localhost:8080/events_for_month?user_id=1&date=2026-04-21'
```

## Принятые решения

- хранилище `in-memory` на `map + sync.RWMutex`
- неделя считается календарной, с понедельника
- месяц считается календарным, от первого до первого числа следующего месяца
- старые события архивируются, если их дата меньше текущей даты
- напоминание считается "отправленным" через структурированный лог

## Тесты

```bash
go test ./...
```
