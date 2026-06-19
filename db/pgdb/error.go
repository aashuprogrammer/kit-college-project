package pgdb

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ErrorNoRow      = "no_rows"
	UniqueViolation = "23505"
)

// ErrorCode returns the database error code if applicable, or "no_rows" for ErrNoRows
func ErrorCode(err error) string {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrorNoRow
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
