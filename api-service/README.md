# API service

Базовый HTTP-сервис магазина на Go. Каркас повторяет подход
`avito-hackathon-8`: конфигурация через окружение, отдельный слой подключения к
PostgreSQL, сборка зависимостей в `main`, router в `internal/handlers` и graceful
shutdown. Вместо GORM используется нативный пул `pgxpool`.

## Запуск

Из корня репозитория:

```bash
cp .env.example .env
make up
```

API будет доступен на `http://localhost:8080`.

```bash
curl http://localhost:8080/api/v1/health/live
curl http://localhost:8080/api/v1/health/ready
open http://localhost:8080/swagger/index.html
```

`live` проверяет процесс приложения, `ready` дополнительно выполняет `Ping`
PostgreSQL.

## Структура

```text
api-service/
  main.go                    # composition root и graceful shutdown
  internal/
    config/                  # env-конфигурация и валидация
    database/                # создание и настройка pgxpool
    handlers/
      router.go              # регистрация маршрутов
      health.go              # liveness/readiness
      cors.go                # HTTP middleware
      response.go            # JSON-ответы
```

Новые домены следует добавлять отдельными пакетами в `internal`, передавать их
сервисам интерфейс хранилища и регистрировать HTTP-обработчики через
`handlers.Dependencies`. SQL остаётся в repository/store-слое доменного пакета,
а транзакции выполняются через `pgx.Tx`.

## Конфигурация

`DATABASE_URL` обязателен. Для Docker Compose он собирается из `POSTGRES_*`; при
локальном запуске его готовое значение приведено в `.env.example`. Redis хранит
ключи идемпотентности и сохранённые HTTP-ответы; время хранения задаётся через
`IDEMPOTENCY_TTL`.
`.env.example` является шаблоном, а не runtime-конфигом: Compose автоматически
читает созданный из него `.env`, а при прямом `go run .` переменные должны быть
экспортированы в окружение процесса.

Основные параметры пула:

```dotenv
DB_MAX_CONNS=25
DB_MIN_CONNS=5
DB_MAX_CONN_LIFETIME=30m
DB_MAX_CONN_IDLE_TIME=5m
DB_HEALTH_CHECK_PERIOD=1m
DB_CONNECT_TIMEOUT=5s
REDIS_URL=redis://localhost:6379/0
IDEMPOTENCY_TTL=24h
```

Проверки:

```bash
go test ./...
go vet ./...
```
