package migrations

import "embed"

// Files contains all Goose SQL migrations.
//
//go:embed *.sql
var Files embed.FS
