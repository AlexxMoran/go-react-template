.DEFAULT_GOAL := help

# Load .env if present so local tooling (migrations) can reuse the same values.
-include .env

# Pinned tool versions (run via `go run`, no global install required).
SQLC_VERSION  := latest
GOOSE_VERSION := latest

# Local database connection (override per environment).
DB_USER     ?= postgres
DB_PASSWORD ?= postgres
DB_NAME     ?= goapp
DB_PORT     ?= 5433
DB_HOST     ?= localhost
DB_SSLMODE  ?= disable
DB_DSN      := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

MIGRATIONS_DIR := db/migrations

## help: show this help
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | awk -F': ' '{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## run: run the API locally
.PHONY: run
run:
	go run ./cmd/api

## build: compile the API binary into ./bin
.PHONY: build
build:
	go build -ldflags="-s -w" -o bin/api ./cmd/api

## test: run all tests
.PHONY: test
test:
	go test ./...

## vet: run go vet
.PHONY: vet
vet:
	go vet ./...

## fmt: format the codebase
.PHONY: fmt
fmt:
	golangci-lint run --fix

## fmt-check: verify formatting without changing files
.PHONY: fmt-check
fmt-check:
	golangci-lint run

## lint: run golangci-lint (requires golangci-lint installed)
.PHONY: lint
lint:
	golangci-lint run

## tidy: tidy go.mod / go.sum
.PHONY: tidy
tidy:
	go mod tidy

## sqlc: regenerate type-safe DB code from db/queries
.PHONY: sqlc
sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

## migrate-up: apply all pending migrations
.PHONY: migrate-up
migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

## migrate-down: roll back the last migration
.PHONY: migrate-down
migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

## migrate-status: show migration status
.PHONY: migrate-status
migrate-status:
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status

## migrate-create: create a new migration (usage: make migrate-create name=add_widgets)
.PHONY: migrate-create
migrate-create:
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir $(MIGRATIONS_DIR) create $(name) sql

## docker-up: build and start the full stack (db + migrations + api)
.PHONY: docker-up
docker-up:
	docker compose up --build

## docker-down: stop the stack and remove volumes
.PHONY: docker-down
docker-down:
	docker compose down -v
