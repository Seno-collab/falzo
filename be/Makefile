migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1

migrate-version:
	migrate -path db/migrations -database "$(DB_URL)" version

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)
