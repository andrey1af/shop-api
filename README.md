# Shop API

REST API магазина бытовой техники на Go.

Сервис позволяет управлять:

- клиентами и их адресами;
- поставщиками;
- товарами и остатками;
- изображениями товаров.

Данные хранятся в PostgreSQL. Redis используется для защиты изменяющих
операций от повторного выполнения.

## Стек

- Go 1.27 и `net/http`;
- PostgreSQL 17;
- Redis 7.4;
- Goose для миграций;
- Docker Compose;
- OpenAPI и Swagger UI.

## Быстрый запуск

Понадобятся Docker и Make.

1. Создайте файл с переменными окружения:

   ```sh
   cp .env.example .env
   ```

2. Запустите приложение:

   ```sh
   make up
   ```

3. Проверьте, что API готово:

   ```sh
   curl http://localhost:8080/api/v1/health/ready
   ```

   Ответ:

   ```json
   {"status":"ok"}
   ```

После запуска доступны:

- API — <http://localhost:8080/api/v1>;
- Swagger UI — <http://localhost:8080/swagger/index.html>;
- OpenAPI-файл — [docs/openapi/v1/api.yaml](docs/openapi/v1/api.yaml).

Миграции применяются автоматически перед запуском API.

## Полезные команды

| Команда | Описание |
|---|---|
| `make up` | Собрать и запустить проект |
| `make down` | Остановить проект |
| `make restart` | Перезапустить сервисы |
| `make logs` | Показать логи |
| `make ps` | Показать состояние контейнеров |
| `make migrate` | Применить миграции |
| `make test` | Запустить тесты |
| `make test-api-service-e2e` | Запустить E2E-тесты |
| `make lint` | Запустить линтер |

## Настройка

Основные параметры находятся в [`.env.example`](.env.example):

```dotenv
POSTGRES_DB=shop
POSTGRES_USER=shop
POSTGRES_PASSWORD=change-me
API_PORT=8080
```

Там же можно настроить таймауты HTTP-сервера, пул соединений PostgreSQL и
время хранения ключей идемпотентности.

## Тесты

Обычные тесты не требуют запущенной базы данных:

```sh
make test
```

E2E-тесты работают с запущенным приложением:

```sh
make up
make test-api-service-e2e
```

## Структура проекта

```text
api-service/       Go-приложение
docs/openapi/      спецификация OpenAPI
migrations/        SQL-миграции
docker-compose.yaml
Makefile
```

## Остановка и удаление данных

Остановить контейнеры:

```sh
make down
```

Удалить контейнеры вместе с локальными данными PostgreSQL и Redis:

```sh
docker compose down --volumes
```
