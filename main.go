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

type Program []Subprogram

func (x Program) String() string {
	var sb strings.Builder
	sb.WriteString("program:\n")
	for i, subprogram := range x {
		if i > 0 {
			sb.WriteString("\n\tthen\n")
		}
		sb.WriteString(subprogram.String())
	}
	return sb.String()
}

type Subprogram []Instruction

func (x Subprogram) String() string {
	var sb strings.Builder
	for i, instr := range x {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(fmt.Sprintf("\t- %s", instr.kind))
		if KIND_MATCH == x[i].kind {
			sb.WriteString(fmt.Sprintf(" %q", instr.payload))
		}
	}
	return sb.String()
}

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

func parseBinaryExpression(query []string, cursor *int, subprogram *Subprogram) error {
	currentWord := query[*cursor]

	if len(*subprogram) < 1 {
		return fmt.Errorf("missing left parameter for binary expression: %s",
			currentWord)
	}
	if len(query)-1 <= *cursor {
		return fmt.Errorf("missing right parameter for binary expression: %q %s",
			(*subprogram)[len(*subprogram)-1].payload, currentWord)
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

	if err := parseExpression(query, cursor, subprogram); err != nil {
		return err
	}
	*subprogram = append(*subprogram, Instruction{
		kind: currentInstructionKind,
	})
	return nil
}

func parseUnaryExpression(query []string, cursor *int, subprogram *Subprogram) error {
	currentWord := query[*cursor]

	switch currentWord {
	case "not":
		if len(query)-1 <= *cursor {
			negation := ""
			if len(*subprogram) > 0 {
				negation = (*subprogram)[len(*subprogram)-1].payload
			}
			return fmt.Errorf("missing right parameter for negation: %q %s",
				negation, currentWord)
		}
		*cursor += 1
		if err := parseExpression(query, cursor, subprogram); err != nil {
			return err
		}
		*subprogram = append(*subprogram, Instruction{
			kind: KIND_NOT,
		})
	default:
		return fmt.Errorf("invalid unary expression: %s", currentWord)
	}

	return nil
}

func parseExpression(query []string, cursor *int, subprogram *Subprogram) error {
	currentWord := query[*cursor]

	switch currentWord {
	case "not":
		if err := parseUnaryExpression(query, cursor, subprogram); err != nil {
			return err
		}
	case "and":
		fallthrough
	case "or":
		if err := parseBinaryExpression(query, cursor, subprogram); err != nil {
			return err
		}
	default:
		if currentWord != "" {
			*subprogram = append(*subprogram, Instruction{
				kind:    KIND_MATCH,
				payload: currentWord,
			})
		}
		*cursor += 1
	}

	return nil
}

func parseSubprogram(query []string) (Subprogram, error) {
	subprogram := make(Subprogram, 0, len(query))
	cursor := 0
	for cursor < len(query) {
		if err := parseExpression(query, &cursor, &subprogram); err != nil {
			return subprogram, err
		}
	}
	return subprogram, nil
}

func ParseProgramSource(programSource string) (Program, error) {
	subqueries := strings.Split(programSource, "|")
	program := make(Program, 0, len(subqueries))
	for _, subquery := range subqueries {
		subprogram, err := parseSubprogram(strings.Split(subquery, " "))
		if err != nil {
			return program, err
		}
		program = append(program, subprogram)
	}
	return program, nil
}

func main() {
	program, err := ParseProgramSource(strings.Join(os.Args[1:], " "))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(program)
	if match, err := MatchesProgram(":bookmark:aws:", program, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("result = %v\n", match)
	}
}
