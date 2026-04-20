# API optimisation

Небольшой HTTP API с endpoint `/sum`, который принимает JSON вида:

```json
{"numbers":[1,2,3]}
```

и возвращает:

```json
{"sum":6,"count":3}
```

## Что сделано

- `cmd/api` поднимает HTTP сервер.
- `internal/api/sumHandlerNaive` показывает базовую реализацию через `io.ReadAll` + `encoding/json`.
- `internal/api/sumHandlerOptimized` использует ручной парсинг строго ожидаемого payload и ручную сборку JSON-ответа, чтобы снизить количество аллокаций и CPU.
- Подключён `net/http/pprof` через `/debug/pprof/*`.
- Добавлены unit-тесты и benchmark-тесты.

## Запуск

```bash
go run ./cmd/api
```

Сервер по умолчанию слушает `:8080`.

Проверка:

```bash
curl -s localhost:8080/sum \
  -H 'Content-Type: application/json' \
  -d '{"numbers":[1,2,3,4]}'
```

## Нагрузка

В репозитории есть простой генератор нагрузки без внешних зависимостей:

```bash
go run ./cmd/loadgen -target http://localhost:8080/sum -c 64 -d 20s
```

Он создаёт конкурентные POST-запросы и печатает:

- общее число успешных запросов;
- число ошибок;
- requests per second;
- latency percentiles `p50`, `p95`, `p99`.

## Benchmark

Запуск benchmark:

```bash
go test -bench . -benchmem ./internal/api
```

Сохранить baseline и сравнить через `benchstat`:

```bash
go test -run '^$' -bench BenchmarkSumHandlerNaive -benchmem ./internal/api > /tmp/naive.txt
go test -run '^$' -bench BenchmarkSumHandlerOptimized -benchmem ./internal/api > /tmp/optimized.txt
benchstat /tmp/naive.txt /tmp/optimized.txt
```

## CPU и memory profiling

Снять CPU profile с живого сервера:

```bash
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```

Снять heap profile:

```bash
go tool pprof http://localhost:8080/debug/pprof/heap
```

Для локального benchmark-профиля:

```bash
go test -run '^$' -bench BenchmarkSumHandlerOptimized \
  -cpuprofile cpu.out \
  -memprofile mem.out \
  ./internal/api
go tool pprof -http=:0 cpu.out
go tool pprof -http=:0 mem.out
```

## Trace

Снять trace для benchmark:

```bash
go test -run '^$' -bench BenchmarkSumHandlerOptimized -trace trace.out ./internal/api
go tool trace trace.out
```

Что смотреть в trace:

- сколько времени тратится внутри handler;
- нет ли лишних блокировок;
- как выглядит распределение времени между `io.ReadAll`, JSON parsing и формированием ответа.

## Какие изменения дали эффект

Базовая версия:

- читает body;
- вызывает `json.Unmarshal` в структуру;
- создаёт response через `json.Marshal`.

Оптимизированная версия:

- парсит только ожидаемый формат `{"numbers":[...]}`;
- не создаёт промежуточные response-структуры;
- использует `strconv.AppendInt` вместо `json.Marshal`;
- переиспользует буфер ответа через `sync.Pool`.

Это снижает число аллокаций и стоимость сериализации на горячем пути.

Локальный benchmark показал:

- `BenchmarkSumHandlerNaive`: `15444 ns/op`, `7543 B/op`, `36 allocs/op`
- `BenchmarkSumHandlerOptimized`: `7051 ns/op`, `6743 B/op`, `22 allocs/op`

То есть оптимизированный путь примерно в 2.2 раза быстрее на этом микробенчмарке и заметно уменьшает количество аллокаций.

## Ограничения

- Оптимизированный parser специально узкий и рассчитан только на ожидаемый формат входа.
- Для real-world API обычно нужен баланс между скоростью, читаемостью и гибкостью схемы.
- `benchstat` должен быть установлен отдельно: `go install golang.org/x/perf/cmd/benchstat@latest`.
