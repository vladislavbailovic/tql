package internal

import (
	"fmt"
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
	case "and":
		fallthrough
	case "or":
		if err := parseBinaryExpression(query, cursor, subprogram); err != nil {
			return err
		}
	case "not":
		if err := parseUnaryExpression(query, cursor, subprogram); err != nil {
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
