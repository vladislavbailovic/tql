package main

import (
	_ "embed"
	"strings"
)

//go:generate go run internal/_buildinfo/main.go commit
//go:embed build/version.txt
var _build_CommitInfo string

//go:generate go run internal/_buildinfo/main.go date
//go:embed build/datetime.txt
var _build_Datetime string

//go:generate go run internal/_buildinfo/main.go compiler
//go:embed build/compiler.txt
var _build_Compiler string

func _build_Info() string {
	var sb strings.Builder

	sb.WriteString("version: ")
	sb.WriteString(_build_CommitInfo)
	sb.WriteByte('\n')
	sb.WriteString(_build_Datetime)
	sb.WriteByte('\n')
	sb.WriteString(_build_Compiler)

	return sb.String()
}
