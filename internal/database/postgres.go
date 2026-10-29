// Package database holds small Postgres helpers shared by all repositories.
package database

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// uniqueViolationCode is the SQLSTATE Postgres returns when an INSERT or
// UPDATE breaks a UNIQUE constraint or unique index.
// Full list: https://www.postgresql.org/docs/current/errcodes-appendix.html
const uniqueViolationCode = "23505"

// IsUniqueViolation reports whether err was caused by a unique constraint.
//
// Repositories insert directly and use this to detect duplicates instead of
// running "SELECT ... then INSERT": two concurrent requests could both pass
// the SELECT, but only one can pass the database's unique index.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	// errors.As looks through wrapped errors for a *pgconn.PgError.
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

// likeEscaper escapes the characters that have special meaning in LIKE/ILIKE
// patterns: "%" (any text), "_" (any single character) and "\" (the escape
// character itself, which must be escaped first).
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// ContainsPattern turns a search term into an ILIKE pattern meaning
// "contains this text literally", e.g. "50%" becomes "%50\%%".
//
// It returns nil for a missing or blank term. Repositories pass the result as
// a SQL parameter and write `$1::text IS NULL OR column ILIKE $1`, so nil
// disables the filter.
func ContainsPattern(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	pattern := "%" + likeEscaper.Replace(strings.TrimSpace(*value)) + "%"
	return &pattern
}
