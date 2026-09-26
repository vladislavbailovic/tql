package main

import (
	"fmt"
	"os"
	"strings"
	"tql/internal"
)

func main() {
	program, err := internal.ParseProgramSource(strings.Join(os.Args[1:], " "))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(program)
	if match, err := internal.MatchesProgram(":bookmark:aws:", program, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("result = %v\n", match)
	}
}
