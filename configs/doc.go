package configs

import "embed"

//go:embed *.yaml
var Configs embed.FS
