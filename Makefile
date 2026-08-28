SHELL := /bin/sh

.DEFAULT_GOAL := help

.PHONY: help up down restart build logs ps migrate test test-api-service test-api-service-e2e lint lint-api-service branch feature bugfix chore

help:
	@printf '%s\n' \
		'make up                   Build and start the application' \
		'make down                 Stop the application' \
		'make restart              Restart the application' \
		'make build                Build all Docker images' \
		'make logs                 Follow service logs' \
		'make ps                   Show service status' \
		'make migrate              Apply database migrations' \
		'make test                 Run all tests' \
		'make test-api-service     Run api-service tests' \
		'make test-api-service-e2e Run supplier E2E tests against the Docker service' \
		'make lint                 Run all linters' \
		'make lint-api-service     Run api-service linter' \
		'make feature NAME=login   Create feature/login' \
		'make bugfix NAME=api      Create bugfix/api' \
		'make chore NAME=deps      Create chore/deps' \
		'make branch TYPE=x NAME=y Create x/y'

up:
	docker compose up --build -d

down:
	docker compose down

restart:
	docker compose restart

build:
	docker compose build

logs:
	docker compose logs --since=1m -f

ps:
	docker compose ps

migrate:
	docker compose run --build --rm migrator

test: test-api-service

test-api-service:
	cd api-service && go test ./...

test-api-service-e2e:
	cd api-service && go test -tags=e2e -count=1 ./test/supplier

lint: lint-api-service

lint-api-service:
	docker run --rm \
		-v "$(CURDIR):/app:ro" \
		-w /app/api-service \
		golangci/golangci-lint:v2.13.1-alpine \
		golangci-lint run --config ../.golangci.yaml

branch:
	@test -n "$(TYPE)" || (printf '%s\n' 'TYPE is required' && exit 1)
	@test -n "$(NAME)" || (printf '%s\n' 'NAME is required' && exit 1)
	@printf '%s' "$(NAME)" | grep -Eq '^[a-z0-9][a-z0-9._-]*$$' || (printf '%s\n' 'NAME must use lowercase letters, numbers, dots, dashes, or underscores' && exit 1)
	git switch -c "$(TYPE)/$(NAME)"

feature:
	@$(MAKE) branch TYPE=feature NAME="$(NAME)"

bugfix:
	@$(MAKE) branch TYPE=bugfix NAME="$(NAME)"

chore:
	@$(MAKE) branch TYPE=chore NAME="$(NAME)"
