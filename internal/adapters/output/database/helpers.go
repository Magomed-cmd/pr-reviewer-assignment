package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgCodeUniqueViolation     = "23505"
	pgCodeForeignKeyViolation = "23503"
)

func isPgError(err error, code string) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}

	return false
}
