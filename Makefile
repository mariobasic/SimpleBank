DB_URL=postgresql://root:password@localhost:5432/bank?sslmode=disable
postgres_up:
	docker-compose up

postgres_down:
	docker-compose down

migrate_up:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrate_up1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migrate_down:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migrate_down1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/mariobasic/simplebank/db/sqlc Store

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
        --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
        --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
        --openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
        proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: postgres_up postgres_down migrate_up migrate_down migrate_up1 migrate_down1 sqlc test server mock db_docs db_schema proto evans

