.DEFAULT_GOAL := help
COMPOSE := docker compose

## help: show this help
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | awk -F': ' '{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## up: build and start the full stack (db + migrations + api + frontend)
.PHONY: up
up: backend/.env
	$(COMPOSE) up --build

## up-d: same as up, detached
.PHONY: up-d
up-d: backend/.env
	$(COMPOSE) up --build -d

## down: stop the stack
.PHONY: down
down:
	$(COMPOSE) down

## clean: stop the stack and remove volumes
.PHONY: clean
clean:
	$(COMPOSE) down -v

## logs: tail all logs
.PHONY: logs
logs:
	$(COMPOSE) logs -f

## backend-%: run a backend Make target (e.g. make backend-test, make backend-sqlc)
.PHONY: backend-%
backend-%:
	$(MAKE) -C backend $*

backend/.env:
	cp backend/.env.example backend/.env
	@echo "Created backend/.env from template — review the JWT secrets before production."
