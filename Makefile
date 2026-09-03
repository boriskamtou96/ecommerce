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
	@echo "  swagger            Regenerate the OpenAPI spec in docs/ from the handler annotations"

include .env
export $(shell test -f .env && sed 's/=.*//' .env)

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
	@echo "Building all binaries...."
	@mkdir -p bin
	@for cmd in cmd/*/; do \
    		if [ -d "$$cmd" ]; then \
    			binary=$$(basename $$cmd); \
    			echo "Building $$binary..."; \
    			go build -o bin/$$binary ./$$cmd; \
    		fi \
    	done

.PHONY: run
run:
	air

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fix-lint
fix-lint:
	golangci-lint run --fix

.PHONY: format
format:
	@gofmt -s -w .
	@goimports -w .

.PHONY: swagger
swagger:
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal --exclude .git,docs,docker,db

.PHONY: up
up:
	docker compose --env-file .env -f docker/docker-compose.yml up -d

.PHONY: down
down:
	docker compose --env-file .env -f docker/docker-compose.yml down

.PHONY: docs-generate
docs-generate:
	mkdir -p docs
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal --exclude .git,docs,docker,db

