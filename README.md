Use sqlc:

1. `docker pull kjconroy/sqlc`
2. `MSYS_NO_PATHCONV=1 docker run --rm -v $(pwd):/src -w /src kjconroy/sqlc generate`
