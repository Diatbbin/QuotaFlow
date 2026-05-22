postgres: 
	docker run --name postgres18 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=test -e POSTGRES_DB=root -d postgres:18-trixie

create-db:
	docker exec -it postgres18 createdb -U root --owner=root QuotaFlow

drop-db:
	docker exec -it postgres18 dropdb QuotaFlow

migrate-up:
	migrate -path db/migration -database "postgres://root:test@localhost:5432/QuotaFlow?sslmode=disable" -verbose up

migrate-down:
	migrate -path db/migration -database "postgres://root:test@localhost:5432/QuotaFlow?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test: test-sqlc test-auth test-util

test-sqlc:
	go test -v -cover -coverpkg=./db/sqlc/... ./db/sqlc_test/... 

test-auth:
	go test -v -cover -coverpkg=./auth/... ./auth/...

test-util:
	go test -v -cover -coverpkg=./util/... ./util/...

server:
	go run main.go

.PHONY: postgres create-db drop-db migrate-up migrate-down sqlc test test-sqlc test-auth test-util server