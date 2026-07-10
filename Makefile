ENV_FILE := .env
COMPOSE := docker compose --env-file $(ENV_FILE)

ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
endif

DB_NAME ?= test
BACKEND_PORT ?= 8080
NGINX_PORT ?= 80

.PHONY: env up down down-v reset-db restart logs logs-backend logs-db logs-nginx logs-prometheus logs-loki logs-grafana ps health db-shell build test test-cover test-docker sqlc mocks lint rebuild route-test

env:
	powershell -NoProfile -Command "if (-not (Test-Path '$(ENV_FILE)')) { Copy-Item '.env.example' '$(ENV_FILE)'; Write-Host 'Created $(ENV_FILE) from .env.example' }"

up: env
	$(COMPOSE) up --build -d --scale backend=2
	@echo Waiting for nginx health...
	powershell -NoProfile -Command "for ($$i=0; $$i -lt 30; $$i++) { try { $$r = Invoke-RestMethod -Uri http://localhost:$(NGINX_PORT)/healthz -TimeoutSec 2; if ($$r.status -eq 'up') { Write-Host 'Backend is up via nginx'; exit 0 } } catch {} Start-Sleep -Seconds 2 }; Write-Host 'Backend not ready yet - check: make logs-nginx'; exit 1"

down:
	$(COMPOSE) down

down-v:
	$(COMPOSE) down -v

reset-db: down-v up

restart:
	$(COMPOSE) restart backend nginx

rebuild: env
	$(COMPOSE) up --build -d --scale backend=2

logs:
	$(COMPOSE) logs -f

logs-backend:
	$(COMPOSE) logs -f backend

logs-nginx:
	$(COMPOSE) logs -f nginx

logs-prometheus:
	$(COMPOSE) logs -f prometheus

logs-loki:
	$(COMPOSE) logs -f loki promtail

logs-grafana:
	$(COMPOSE) logs -f grafana

logs-db:
	$(COMPOSE) logs -f postgres_db

ps:
	$(COMPOSE) ps

health:
	powershell -NoProfile -Command "Invoke-RestMethod -Uri http://localhost:$(NGINX_PORT)/healthz | ConvertTo-Json"

route-test:
	powershell -NoProfile -Command "1..10 | ForEach-Object { Invoke-RestMethod -Uri http://localhost:$(NGINX_PORT)/healthz | Out-Null }; Write-Host 'Sent 10 requests. Check replica routing with: make logs-nginx'"

db-shell:
	docker exec -it keeneye_postgres psql -U postgres -d $(DB_NAME)

build:
	cd backend && go build -o bin/server ./app/cmd/server/main.go

test:
	cd backend && go test ./... -count=1

test-cover:
	cd backend && go test ./... -coverprofile=coverage.out

test-docker:
	docker run --rm -v "$(CURDIR)/backend:/app" -w /app golang:1.26-alpine sh -c "apk add --no-cache git >/dev/null && go test ./... -count=1"

sqlc:
	cd backend && sqlc generate

mocks:
	cd backend && go run github.com/vektra/mockery/v2@v2.53.6

lint:
	cd backend && go vet ./...
