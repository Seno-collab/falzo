
run:
	go run cmd/api/main.go

test:
	go test ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down
