package auth

// Credential rules shared by every account type (users and customers).
// Keeping them here means both gates always apply the same rules, and the
// login request in auth_model.go stays consistent with account creation.

import (
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/ladev707/3dbuilder-backend/internal/apperr"
)

// NormalizeUsername removes surrounding whitespace so " jane " and "jane"
// are stored identically.
func NormalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

// NormalizeEmail trims and lowercases an email address. Login already
// compares emails case-insensitively; storing them lowercased keeps the data
// consistent too.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateUsername requires 3 to 80 characters. The upper limit matches the
// VARCHAR(80) column in migrations/00001_auth_rbac.sql.
func ValidateUsername(username string) error {
	// RuneCountInString counts characters, not bytes, so "jöhn" is 4 long.
	if length := utf8.RuneCountInString(username); length < 3 || length > 80 {
		return apperr.New(apperr.InvalidInput, "username must be between 3 and 80 characters")
	}
	return nil
}

// ValidateEmail accepts a bare address such as "jane@example.com".
//
// mail.ParseAddress also accepts forms like `"Jane" <jane@example.com>`, so
// the parsed address must equal the input exactly. The 255 limit matches the
// VARCHAR(255) column.
func ValidateEmail(email string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 255 {
		return apperr.New(apperr.InvalidInput, "email must be a valid email address")
	}
	return nil
}

// ValidatePassword requires 8 to 72 bytes. bcrypt rejects passwords longer
// than 72 bytes, so the limit is checked before hashing.
func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return apperr.New(apperr.InvalidInput, "password must be between 8 and 72 bytes")
	}
	return nil
}
