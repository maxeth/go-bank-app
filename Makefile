postgres:
	docker run --name postgrestut -p 5433:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -u postgres postgrestut createdb --username=postgres --owner=postgres bank_app

dropdb:
	docker stop postgrestut && docker start postgrestut && docker exec -u postgres postgrestut dropdb bank_app

migrateup:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:5433/bank_app?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:5433/bank_app?sslmode=disable" -verbose up

sqlc-gen:
	MSYS_NO_PATHCONV=1 docker run --rm -v $(pwd):/src -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown