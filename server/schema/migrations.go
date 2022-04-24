package schema

import "embed"

//go:embed *.sql
var MigrationFiles embed.FS

const Label = "user"
