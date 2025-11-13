DB_URL := "postgres://postgres:password@127.0.0.1:5433/pr_reviewer_db?sslmode=disable"

.PHONY: migrate-up migrate-down migrate-force

migrate-up:
	migrate -path migrations -database $(DB_URL) up

migrate-down:
	migrate -path migrations -database $(DB_URL) down

migrate-force:
	migrate -path migrations -database $(DB_URL) force $(VERSION)
