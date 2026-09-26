package internal

import (
	"fmt"
	"strings"
)

func MatchesProgram(subject string, program Program, backtrace *[]string) (bool, error) {
	for _, subprogram := range program {
		if result, err := matchSubprogram(subject, subprogram, backtrace); err != nil {
			return result, err
		} else if result != true {
			return false, nil
		}
	}
	return true, nil
}

func matchSubprogram(subject string, subprogram Subprogram, backtrace *[]string) (bool, error) {
	if backtrace != nil {
		*backtrace = append(*backtrace, fmt.Sprintf("Matching against %q", subject))
	}
	stack := make([]bool, 0, len(subprogram))

	for i := 0; i < len(subprogram); i++ {
		switch subprogram[i].kind {
		case KIND_MATCH:
			value := false
			if strings.Contains(subject, fmt.Sprintf(":%s:", subprogram[i].payload)) {
				value = true
			}
			stack = append(stack, value)
			if backtrace != nil {
				*backtrace = append(*backtrace,
					fmt.Sprintf("\t- MATCH %q: %v", subprogram[i].payload, value))
			}
		case KIND_AND:
			if len(stack) < 2 {
				return false, fmt.Errorf("%s:%d expects 2 values, got %d",
					subprogram[i].kind, i, len(stack))
			}
			left := stack[len(stack)-2]
			right := stack[len(stack)-1]
			value := left && right
			stack[len(stack)-2] = value
			stack = stack[:len(stack)-1]
			if backtrace != nil {
				*backtrace = append(*backtrace,
					fmt.Sprintf("\t- %v AND %v: %v", left, right, value))
			}
		case KIND_OR:
			if len(stack) < 2 {
				return false, fmt.Errorf("%s:%d expects 2 values, got %d",
					subprogram[i].kind, i, len(stack))
			}
			left := stack[len(stack)-2]
			right := stack[len(stack)-1]
			value := left || right
			stack[len(stack)-2] = value
			stack = stack[:len(stack)-1]
			if backtrace != nil {
				*backtrace = append(*backtrace,
					fmt.Sprintf("\t- %v OR %v: %v", left, right, value))
			}
		case KIND_NOT:
			if len(stack) < 1 {
				return false, fmt.Errorf("%s:%d expects 1 value, got %d",
					subprogram[i].kind, i, len(stack))
			}
			last := stack[len(stack)-1]
			stack[len(stack)-1] = !last
			if backtrace != nil {
				*backtrace = append(*backtrace,
					fmt.Sprintf("\t- NOT %v: %v", last, !last))
			}
		case KIND_INVALID:
			return false, fmt.Errorf("invalid instruction at %d", i)
		default:
			return false, fmt.Errorf("match not implemented: %d at %d",
				subprogram[i].kind, i)
		}
	}

	if len(stack) != 1 {
		return false, fmt.Errorf("ambiguous result, got %d values", len(stack))
	}
	if backtrace != nil {
		*backtrace = append(*backtrace,
			fmt.Sprintf("result = %v", stack[0]))
	}

	return stack[0], nil
}
