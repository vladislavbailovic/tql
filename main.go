package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"tql/internal"
)

func main() {
	program, err := internal.ParseProgramSource(strings.Join(os.Args[1:], " "))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if home, err := os.UserHomeDir(); err == nil {
		files, _ := filepath.Glob(path.Join(home, "vimwiki", "**/*.md"))
		for _, file := range files {
			matchFile(file, program)
		}
	}
}

func matchFile(file string, program internal.Program) error {
	abspath, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(abspath)
	if err != nil {
		return err
	}

	var sb strings.Builder
	for i, line := range strings.Split(string(content), "\n") {
		if match, _ := internal.MatchesProgram(line, program, nil); match {
			sb.WriteString(fmt.Sprintf(
				"%s:%d: %s\n", file, i, line))
		}
	}
	fmt.Print(sb.String())
	return nil
}
