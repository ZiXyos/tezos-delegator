package conf

import "embed"

//go:embed *.toml
var FileFS embed.FS
