postgres_up:
	docker-compose up

postgres_down:
	docker-compose down

migrate_up:
	migrate -path db/migration -database "postgresql://root:password@localhost:5432/bank?sslmode=disable" -verbose up

migrate_down:
	migrate -path db/migration -database "postgresql://root:password@localhost:5432/bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres_up postgres_down migrate_up migrate_down sqlc test

