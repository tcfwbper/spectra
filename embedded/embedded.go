package embedded

import "embed"

//go:embed roles/*
var RolesFS embed.FS

//go:embed skills/*
var SkillsFS embed.FS

//go:embed README.md
var ReadmeFS embed.FS

//go:embed all:templates
var TemplatesFS embed.FS
