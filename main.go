package main

import (
	"fmt"
	"os"
	"strings"
)

type InstructionKind uint8

func (x InstructionKind) String() string {
	switch x {
	case KIND_INVALID:
		return "INVALID"
	case KIND_MATCH:
		return "MATCH"
	case KIND_AND:
		return "AND"
	case KIND_OR:
		return "OR"
	case KIND_NOT:
		return "NOT"
	default:
		return fmt.Sprintf("Unknown instruction kind: %d", int(x))
	}
}

const (
	KIND_INVALID InstructionKind = iota
	KIND_MATCH   InstructionKind = iota
	KIND_AND     InstructionKind = iota
	KIND_OR      InstructionKind = iota
	KIND_NOT     InstructionKind = iota
)

type Instruction struct {
	kind    InstructionKind
	payload string
}

func printProgram(program []Instruction) {
	fmt.Printf("program:\n")
	for i := 0; i < len(program); i++ {
		fmt.Printf("\t- %s", program[i].kind)
		if KIND_MATCH == program[i].kind {
			fmt.Printf(" %q", program[i].payload)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}

func match(subject string, program []Instruction, backtrace *[]string) (bool, error) {
	if backtrace != nil {
		*backtrace = append(*backtrace, fmt.Sprintf("Matching against %q", subject))
	}
	stack := make([]bool, 0, len(program))

	for i := 0; i < len(program); i++ {
		switch program[i].kind {
		case KIND_MATCH:
			value := false
			if strings.Contains(subject, fmt.Sprintf(":%s:", program[i].payload)) {
				value = true
			}
			stack = append(stack, value)
			if backtrace != nil {
				*backtrace = append(*backtrace,
					fmt.Sprintf("\t- MATCH %q: %v", program[i].payload, value))
			}
		case KIND_AND:
			if len(stack) < 2 {
				return false, fmt.Errorf("%s:%d expects 2 values, got %d",
					program[i].kind, i, len(stack))
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
					program[i].kind, i, len(stack))
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
					program[i].kind, i, len(stack))
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
				program[i].kind, i)
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

func parseBinaryExpression(query []string, cursor *int, program *[]Instruction) error {
	currentWord := query[*cursor]

	if len(*program) < 1 {
		return fmt.Errorf("missing left parameter for binary expression: %s",
			currentWord)
	}
	if len(query)-1 <= *cursor {
		return fmt.Errorf("missing right parameter for binary expression: %q %s",
			(*program)[len(*program)-1].payload, currentWord)
	}

	var currentInstructionKind InstructionKind
	switch currentWord {
	case "and":
		currentInstructionKind = KIND_AND
		*cursor += 1
	case "or":
		currentInstructionKind = KIND_OR
		*cursor += 1
	default:
		return fmt.Errorf("invalid binary expression: %s", currentWord)
	}

	if err := parseExoression(query, cursor, program); err != nil {
		return err
	}
	*program = append(*program, Instruction{
		kind: currentInstructionKind,
	})
	return nil
}

func parseUnaryExpression(query []string, cursor *int, program *[]Instruction) error {
	currentWord := query[*cursor]

	switch currentWord {
	case "not":
		if len(query)-1 <= *cursor {
			negation := ""
			if len(*program) > 0 {
				negation = (*program)[len(*program)-1].payload
			}
			return fmt.Errorf("missing right parameter for negation: %q %s",
				negation, currentWord)
		}
		*cursor += 1
		if err := parseExoression(query, cursor, program); err != nil {
			return err
		}
		*program = append(*program, Instruction{
			kind: KIND_NOT,
		})
	default:
		return fmt.Errorf("invalid unary expression: %s", currentWord)
	}

	return nil
}

func parseExoression(query []string, cursor *int, program *[]Instruction) error {
	currentWord := query[*cursor]

	switch currentWord {
	case "not":
		if err := parseUnaryExpression(query, cursor, program); err != nil {
			return err
		}
	case "and":
		fallthrough
	case "or":
		if err := parseBinaryExpression(query, cursor, program); err != nil {
			return err
		}
	default:
		*program = append(*program, Instruction{
			kind:    KIND_MATCH,
			payload: currentWord,
		})
		*cursor += 1
	}

	return nil
}

func parseQuery(query []string) ([]Instruction, error) {
	program := make([]Instruction, 0, len(query))
	cursor := 0
	for cursor < len(query) {
		if err := parseExoression(query, &cursor, &program); err != nil {
			return program, err
		}
	}
	return program, nil
}

func parseQueryString(queryString string) ([]Instruction, error) {
	query := strings.Split(queryString, " ")
	return parseQuery(query)
}

func main() {
	program, err := parseQuery(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printProgram(program)
	if match, err := match(":bookmark:aws:", program, nil); err != nil {
		printProgram(program)
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("result = %v\n", match)
	}
}
