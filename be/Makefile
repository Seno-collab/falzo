-include .env

DB_URL ?= postgres://$(DATABASE_USER_NAME):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)

.PHONY: migrate-up migrate-down migrate-version migrate-create migrate-reset

migrate-up:
	@migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	@migrate -path migrations -database "$(DB_URL)" down 1

migrate-version:
	@migrate -path migrations -database "$(DB_URL)" version

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-reset:
	@if [ "$(CONFIRM)" != "yes" ]; then \
		echo "Refusing to reset database. Re-run with CONFIRM=yes"; \
		exit 1; \
	fi
	@migrate -path migrations -database "$(DB_URL)" drop -f
	@migrate -path migrations -database "$(DB_URL)" up
run:
	go run cmd/api/main.go
