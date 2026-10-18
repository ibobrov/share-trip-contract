.DEFAULT_GOAL := help

EXE :=
ifeq ($(OS),Windows_NT)
EXE := .exe
endif

export GOBIN := $(CURDIR)/.tools

LINT_VERSION := v2.11.3
GOOSE_VERSION := v3.28.0
OAPI_CODEGEN_VERSION := v2.8.0

OAPI_CODEGEN_BIN := $(GOBIN)/oapi-codegen$(EXE)

LINT := $(GOBIN)/golangci-lint$(EXE)
GOOSE := $(GOBIN)/goose$(EXE)

COMPOSE := docker compose --env-file .env -f deploy/docker-compose.yml

.PHONY: help deps tidy fmt lint test build run up down
.PHONY: migrate-create migrate-status migrate-up migrate-down
.PHONY: generate

help:
	@echo Available commands:
	@echo   make deps               - Download modules and install pinned CLI tools
	@echo   make tidy               - Update go.mod/go.sum and remove unused requirements
	@echo   make fmt                - Format Go code
	@echo   make lint               - Run linters
	@echo   make test               - Run all tests
	@echo   make build              - Generate API and build bin/sharetrip-contract
	@echo   make run                - Run application
	@echo   make generate           - Generate API code
	@echo   make up                 - Start local PostgreSQL and wait for health
	@echo   make down               - Stop PostgreSQL and preserve its volume
	@echo   make migrate-create name=create_contracts - Create SQL migration
	@echo   make migrate-status     - Show migration status
	@echo   make migrate-up         - Apply migrations
	@echo   make migrate-down       - Roll back last migration with possible data loss
	@echo   make help               - Show available commands

deps: $(LINT) $(GOOSE) $(OAPI_CODEGEN_BIN)
	go mod download

$(OAPI_CODEGEN_BIN):
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)

generate: $(OAPI_CODEGEN_BIN)
	"$(OAPI_CODEGEN_BIN)" --config api/openapi.codegen.yaml api/contract.yaml

$(LINT):
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINT_VERSION)

$(GOOSE):
	go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

tidy:
	go mod tidy

fmt:
	go fmt ./...

lint: $(LINT)
	"$(LINT)" run ./...

test:
	go test ./...

build: generate
	go build -o bin/sharetrip-contract$(EXE) ./cmd/contract

run: generate
	go run ./cmd/contract

up:
	$(COMPOSE) up -d --wait --wait-timeout 90

down:
	$(COMPOSE) down

migrate-create: $(GOOSE)
ifeq ($(OS),Windows_NT)
	powershell.exe -NoProfile -Command "New-Item -ItemType Directory -Force -Path 'migrations' | Out-Null"
else
	mkdir -p migrations
endif
	"$(GOOSE)" -env none -dir migrations create "$(if $(strip $(name)),$(strip $(name)),name)" sql

migrate-status: $(GOOSE)
	"$(GOOSE)" -env .env -dir migrations status

migrate-up: $(GOOSE)
	"$(GOOSE)" -env .env -dir migrations up

migrate-down: $(GOOSE)
	"$(GOOSE)" -env .env -dir migrations down
