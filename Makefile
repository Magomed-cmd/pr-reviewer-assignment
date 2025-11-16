DB_URL := "postgres://postgres:password@127.0.0.1:5433/pr_reviewer_db?sslmode=disable"

.PHONY: migrate-up migrate-down migrate-force run lint test test-integration test-e2e docker-up docker-down

migrate-up:
	migrate -path migrations -database $(DB_URL) up

migrate-down:
	migrate -path migrations -database $(DB_URL) down

migrate-force:
	migrate -path migrations -database $(DB_URL) force $(VERSION)

run:
	go run cmd/server/main.go

lint:
	go vet ./...
	golangci-lint run

test:
	go test ./...

test-integration:
	go test -p 1 -count=1 ./tests/integration

test-e2e:
	go test -p 1 -count=1 ./tests/e2e

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v
