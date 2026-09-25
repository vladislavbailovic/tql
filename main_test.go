package main

import (
	"fmt"
	"strings"
	"testing"
)

func Test_ParseErrors(t *testing.T) {
	suite := []struct {
		query string
		fail  bool
	}{
		{query: "bookmark and aws", fail: false},
		{query: "not work", fail: false},
		{query: "not work or aws", fail: false},

		{query: "bookmark and", fail: true},
		{query: "work not", fail: true},
		{query: "or work", fail: true},
		{query: "not", fail: true},
	}
	for _, test := range suite {
		t.Run(test.query, func(t *testing.T) {
			program, err := ParseProgramSource(test.query)
			if test.fail && err == nil {
				t.Errorf("expected error, but didn't get one:\n%s", program)
			}
			if !test.fail && err != nil {
				t.Errorf("error parsing %q: %v\n%s", test.query, err, program)
			}
		})
	}
}

func Test_MatchHappyPath(t *testing.T) {
	suite := []struct {
		query    string
		line     string
		expected bool
	}{
		{
			query:    "not work",
			line:     ":bookmark:aws:",
			expected: true,
		},
		{
			query:    "bookmark and not work",
			line:     ":bookmark:aws:",
			expected: true,
		},
		{
			query:    "bookmark and aws",
			line:     ":bookmark:aws:",
			expected: true,
		},
		{
			query:    "bookmark and aws",
			line:     ":bookmark:aws",
			expected: false,
		},
		{
			query:    "bookmark and aws",
			line:     "bookmark:aws:",
			expected: false,
		},
	}
	for _, test := range suite {
		t.Run(test.query, func(t *testing.T) {
			program, err := ParseProgramSource(test.query)
			if err != nil {
				t.Fatal(err)
			}

			backtrace := make([]string, 0, len(program))
			if res, err := MatchesProgram(test.line, program, &backtrace); err != nil {
				fmt.Printf("%s\nBacktrace %q:\n%s\n",
					program, test.query, strings.Join(backtrace, "\n"))
				t.Errorf("error matching query %q against %q: %v",
					test.query, test.line, err)
			} else {
				if test.expected != res {
					fmt.Printf("%s\nBacktrace %q:\n%s\n",
						program, test.query, strings.Join(backtrace, "\n"))
					t.Errorf("match query %q against %q: want %v, got %v",
						test.query, test.line, test.expected, res)
				}
			}
		})
	}
}
