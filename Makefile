
run-be:
	cd be && if [ -f .env ]; then set -a && . ./.env && set +a; fi; go run ./cmd/api

up-db:
	docker compose up -d postgres

down-db:
	docker compose down

test-api-auth:
	./scripts/test-auth-api.sh
