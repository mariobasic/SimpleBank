package db

import (
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ForeignKeyViolationCode = "23503"
	UniqueViolationCode     = "23505"
)

var ErrRecordNotFound = pgx.ErrNoRows
var ErrUniqueViolation = &pgconn.PgError{Code: UniqueViolationCode, Message: "unique violation"}

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
