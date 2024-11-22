include .envrc
MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: test
test:
	@go test -v ./...

.PHONY: migration
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: up-migrate
up-migrate:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDRESS) -verbose up

.PHONY: down-migrate
down-migrate:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDRESS) -verbose down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: seed
seed: 
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

docker_up:
	docker compose up -d

docker_down:
	docker compose down

create_db:
	docker exec -it postgres-db createdb --username=admin --owner=root simple_bank

drop_db:
	docker exec -it postgres-db dropdb simple_bank

fmt:
	go fmt ./...