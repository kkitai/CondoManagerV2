GOENV_GO := $(HOME)/.goenv/versions/$(shell cat $(HOME)/.goenv/version 2>/dev/null || echo 1.26.2)/bin/go
GO ?= $(shell command -v go 2>/dev/null || echo $(GOENV_GO))
GOOSE ?= $(shell command -v goose 2>/dev/null || echo goose)

DB_DSN ?= "host=localhost port=5432 user=postgres password=postgres dbname=condo_manager sslmode=disable"
MIGRATIONS_DIR := app/db/migrations
BINARY := bin/server

.PHONY: all build run test test-coverage migrate migrate-down migrate-status clean docker-up docker-down vet lint

all: build

build:
	$(GO) build -o $(BINARY) ./app/cmd/server/

run:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	$(GO) run ./app/cmd/server/

test:
	$(GO) test ./... -count=1

test-coverage:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html
	$(GO) tool cover -func=coverage.out | tail -1

test-coverage-check:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	@total=$$($(GO) tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	echo "Coverage: $$total%"; \
	if [ $$(echo "$$total < 70" | bc -l) -eq 1 ]; then \
		echo "Coverage below 70%"; exit 1; \
	fi

migrate:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) status

migrate-reset:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) reset

vet:
	$(GO) vet ./...

clean:
	rm -f $(BINARY) coverage.out coverage.html

docker-up:
	docker compose up -d

docker-down:
	docker compose down
