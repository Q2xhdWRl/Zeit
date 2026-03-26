package migrations

import "embed"

//go:embed *.up.sql
var FS embed.FS

//go:embed dev_seed.sql
var DevSeedSQL string
