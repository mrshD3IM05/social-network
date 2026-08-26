package migrations

import "embed"

//go:embed sqlite/*.sql
var FS embed.FS
