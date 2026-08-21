.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  migration <name>   Create a new migration with the given name"
	@echo "  migrate-up         Apply all up migrations"
	@echo "  migrate-down <n>   Rollback the last n migrations"
	@echo "  build              Build the application"
	@echo "  run                Run the application"
	@echo "  lint               Run golangci-lint on the codebase"
	@echo "  format             Format code an re-arrange imports"

# include .envrc
MIGRATIONS_PATH = ./db/migrations
DB_ADDR ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir ${MIGRATIONS_PATH} $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=${MIGRATIONS_PATH} -database=$(DB_ADDR) up


.PHONY: migrate-down
migrate-down:
	@migrate -path=${MIGRATIONS_PATH} -database=$(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: build
build:
	go build -o bin/app ./cmd/api

.PHONY: run
run:
	air

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: format
format:
	@gofmt -s -w .
	@goimports -w .

.PHONY: up
up:
	docker compose -f docker/docker-compose.yml up -d

.PHONY: down
down:
	docker compose -f docker/docker-compose.yml down

