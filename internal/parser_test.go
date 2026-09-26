package internal

import "testing"

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
