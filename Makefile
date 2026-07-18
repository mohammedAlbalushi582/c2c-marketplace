# Convenience targets. Migrations & sqlc run via Docker images (no local install needed).

DB_URL ?= postgres://alamjad:alamjad_dev_pw@localhost:5432/alamjad?sslmode=disable
MIGRATE = docker run --rm -v $(PWD)/migrations:/migrations --network host migrate/migrate:v4.18.1 \
          -path=/migrations -database "$(DB_URL)"

.PHONY: up down logs db-up migrate-up migrate-down migrate-force sqlc new-migration backend-run

## Infra
up:            ## start postgres (detached)
	docker compose up -d postgres

down:          ## stop all services
	docker compose down

logs:          ## tail compose logs
	docker compose logs -f

## Migrations (golang-migrate via Docker)
migrate-up:    ## apply all up migrations
	$(MIGRATE) up

migrate-down:  ## roll back one migration
	$(MIGRATE) down 1

migrate-force: ## force a version (usage: make migrate-force V=3)
	$(MIGRATE) force $(V)

new-migration: ## scaffold a migration (usage: make new-migration NAME=add_x)
	docker run --rm -v $(PWD)/migrations:/migrations migrate/migrate:v4.18.1 \
		create -ext sql -dir /migrations -seq $(NAME)

## Codegen
sqlc:          ## regenerate type-safe queries
	docker run --rm -v $(PWD):/work -w /work/backend sqlc/sqlc:1.27.0 generate

## Backend
backend-run:   ## run the API locally (needs .env + running postgres)
	cd backend && go run ./cmd/api
