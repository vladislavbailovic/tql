package main

import (
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
			program, err := parseQueryString(test.query)
			if test.fail && err == nil {
				printProgram(program)
				t.Error("expected error, but didn't get one")
			}
			if !test.fail && err != nil {
				printProgram(program)
				t.Errorf("error parsing %q: %v", test.query, err)
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
			program, err := parseQueryString(test.query)
			if err != nil {
				t.Fatal(err)
			}

			if res, err := match(test.line, program); err != nil {
				printProgram(program)
				t.Errorf("error matching query %q against %q: %v",
					test.query, test.line, err)
			} else {
				if test.expected != res {
					t.Errorf("match query %q against %q: want %v, got %v",
						test.query, test.line, test.expected, res)
				}
			}
		})
	}
}
