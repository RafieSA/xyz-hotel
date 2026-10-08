.PHONY: run migrate tidy vet build

run:
	go run ./cmd/server

migrate:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

tidy:
	go mod tidy

vet:
	go vet ./...

build:
	go build -o bin/server ./cmd/server
