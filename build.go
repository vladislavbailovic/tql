package main

import (
	_ "embed"
	"strings"
)

//go:generate sh -c "git describe HEAD --tags --always > build/version.txt"
//go:embed build/version.txt
var _build_CommitInfo string

//go:generate sh -c "date > build/datetime.txt"
//go:embed build/datetime.txt
var _build_Date string

//go:generate sh -c "go version > build/compiler.txt"
//go:embed build/compiler.txt
var _build_Compiler string

func _build_Info() string {
	var sb strings.Builder

	sb.WriteString("version: ")
	sb.WriteString(_build_CommitInfo)
	sb.WriteString(_build_Date)
	sb.WriteString(_build_Compiler)

	return sb.String()
}
