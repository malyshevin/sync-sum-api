.PHONY: run migrate-up migrate-down tidy build compose-up compose-down

run:
	go run ./cmd/server

migrate-up:
	go run ./cmd/migrate -dir $${MIGRATIONS_DIR:-migrations}

migrate-down:
	go run ./cmd/migrate -version 1 -dir $${MIGRATIONS_DIR:-migrations}

tidy:
	go mod tidy

build:
	go build ./...

compose-up:
	docker compose up -d postgres

compose-down:
	docker compose down


