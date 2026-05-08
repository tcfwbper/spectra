package builtin

import "embed"

//go:embed workflows/*.yaml
var Workflows embed.FS

//go:embed agents/*.yaml
var Agents embed.FS

//go:embed spec/ARCHITECTURE.md
//go:embed spec/CONVENTIONS.md
//go:embed spec/logic/README.md
//go:embed spec/test/README.md
var SpecFiles embed.FS
