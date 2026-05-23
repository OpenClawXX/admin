.PHONY: run build migrate-up migrate-down

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

migrate-up:
	goose -dir migrations postgres "$$(grep DB_ .env | tr '\n' ' ' | sed 's/DB_HOST=\([^ ]*\) DB_PORT=\([^ ]*\) DB_USER=\([^ ]*\) DB_PASSWORD=\([^ ]*\) DB_NAME=\([^ ]*\) DB_SSLMODE=\([^ ]*\)/host=\1 port=\2 user=\3 password=\4 dbname=\5 sslmode=\6/')" up

migrate-down:
	goose -dir migrations postgres "$$(grep DB_ .env | tr '\n' ' ' | sed 's/DB_HOST=\([^ ]*\) DB_PORT=\([^ ]*\) DB_USER=\([^ ]*\) DB_PASSWORD=\([^ ]*\) DB_NAME=\([^ ]*\) DB_SSLMODE=\([^ ]*\)/host=\1 port=\2 user=\3 password=\4 dbname=\5 sslmode=\6/')" down

test:
	go test ./... -v
