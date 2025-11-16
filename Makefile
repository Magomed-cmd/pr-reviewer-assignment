DB_URL := "postgres://postgres:password@127.0.0.1:5433/pr_reviewer_db?sslmode=disable"

.PHONY: migrate-up migrate-down migrate-force run lint test test-integration test-e2e docker-up docker-down

BASE_URL ?= http://host.docker.internal:8080

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

load-test:
	@if command -v k6 >/dev/null 2>&1; then \
		BASE_URL=$(BASE_URL) k6 run ./scripts/load_test/k6.js; \
	else \
		echo "k6 not found, running via docker image grafana/k6"; \
		docker run --rm -i -v "$${PWD}":/scripts grafana/k6 run /scripts/scripts/load_test/k6.js -e BASE_URL=$(BASE_URL); \
	fi
