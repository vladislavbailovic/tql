package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"tql/internal"
)

var commands = []string{
	"ls",
	"print",
}

func PrintCommandHelp() {
	fmt.Println("Supported commands:")
	fmt.Println(strings.Join(commands, ", "))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing subcommand")
		PrintCommandHelp()
		os.Exit(1)
	}
	var cmd string
	for _, x := range commands {
		if x == os.Args[1] {
			cmd = x
			break
		}
	}
	if cmd == "" {
		fmt.Printf("Unknown subcommand %q\n", os.Args[1])
		PrintCommandHelp()
		os.Exit(1)
	}

	args := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	help := args.Bool("help", false, "Show help")
	version := args.Bool("version", false, "Show version information")
	// TODO: move source matching into TQL proper
	// This would work by introducing an instructionkind for source
	source := args.String("source", "", "Filter file basenames")
	args.Parse(os.Args[2:])

	if *help {
		args.Usage()
		os.Exit(0)
	}
	if *version {
		fmt.Println(_build_Info())
		os.Exit(0)
	}

	program, err := internal.ParseProgramSource(strings.Join(args.Args(), " "))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if home, err := os.UserHomeDir(); err == nil {
		files, _ := filepath.Glob(path.Join(home, "vimwiki", "**/*.md"))
		for _, file := range files {
			if *source != "" && !strings.HasPrefix(filepath.Base(file), *source) {
				continue
			}
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
