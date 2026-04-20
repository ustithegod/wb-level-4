# GC Analyzer

Небольшой HTTP-сервер на Go, который публикует в формате Prometheus текущие метрики памяти и сборщика мусора Go runtime.

Что реализовано:

- endpoint `/metrics` с данными из `runtime.ReadMemStats`
- настройка GC через `debug.SetGCPercent`
- профилирование через `pprof`
- endpoint `/healthz` для проверки доступности сервиса

## Запуск

Требования:

- Go 1.25+

Команда запуска:

```bash
go run ./cmd/gc-analyzer
```

Опциональные переменные окружения:

- `ADDR` — адрес HTTP-сервера, по умолчанию `:8080`
- `GC_PERCENT` — значение для `debug.SetGCPercent`, по умолчанию `100`

Пример:

```bash
ADDR=:9090 GC_PERCENT=200 go run ./cmd/gc-analyzer
```

## Endpoint'ы

- `GET /metrics`
- `GET /healthz`
- `GET /debug/pprof/`
- `GET /debug/pprof/heap`
- `GET /debug/pprof/profile?seconds=10`

## Примеры запросов

Получить метрики:

```bash
curl http://localhost:8080/metrics
```

Проверить здоровье сервиса:

```bash
curl http://localhost:8080/healthz
```

Снять CPU profile на 10 секунд:

```bash
curl http://localhost:8080/debug/pprof/profile?seconds=10 --output cpu.pprof
```

Посмотреть heap profile:

```bash
go tool pprof http://localhost:8080/debug/pprof/heap
```

## Примеры метрик

- `go_gc_mallocs_total`
- `go_gc_cycles_total`
- `go_memory_alloc_bytes`
- `go_gc_last_run_time_seconds`
- `go_runtime_goroutines`
- `go_gc_target_percent`
