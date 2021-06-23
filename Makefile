postgres:
	docker run --name postgrestut -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -u postgres postgrestut createdb --username=postgres --owner=postgres bank_app

dropdb:
	docker stop postgrestut && docker start postgrestut && docker exec -u postgres postgrestut dropdb bank_app

migrateup:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:5432/bank_app?sslmode=disable" -verbose up

migrateupone:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:5432/bank_app?sslmode=disable" -verbose up 1
 
migratedown:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:5432/bank_app?sslmode=disable" -verbose down

migratedownone:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:5432/bank_app?sslmode=disable" -verbose down 1

sqlc-gen:
	MSYS_NO_PATHCONV=1 docker run --rm -v $(pwd):/src -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mockdb: 
	mockgen -package mockdb -destination ./db/mock/repository.go github.com/maxeth/go-bank-app/db/sqlc Repository


.PHONY: postgres createdb dropdb migrateup migratedown server mock-db sqlc-gen migrateupone migratedownone test
