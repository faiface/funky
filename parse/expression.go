package parse

import (
	"fmt"

	"github.com/faiface/funky/expr"
)

type Error struct {
	SourceInfo *expr.SourceInfo
	Msg        string
}

func (pe *Error) Error() string {
	return fmt.Sprintf("%v: %v", pe.SourceInfo, pe.Msg)
}

func Expression(tokens []Token) (expr.Expr, error) {
	var e expr.Expr

	for len(tokens) > 0 {
		switch tokens[0].Value {
		case "[", "]", "{", "}", ",":
			return nil, &Error{
				tokens[0].SourceInfo,
				fmt.Sprintf("invalid character: %s", tokens[0].Value),
			}

		case ")":
			return nil, &Error{
				tokens[0].SourceInfo,
				"no matching opening parenthesis",
			}

		case "(":
			closing := findClosingParen(tokens)
			if closing == -1 {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no matching closing parenthesis",
				}
			}
			if closing == 1 {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no expression in parentheses",
				}
			}
			inParen, err := Expression(tokens[1:closing])
			if err != nil {
				return nil, err
			}
			e = wrapAppl(e, inParen)
			tokens = tokens[closing+1:]

		case "\\", "λ":
			if len(tokens) < 2 {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no bound variable in abstraction",
				}
			}
			if IsSpecial(tokens[1].Value) {
				return nil, &Error{
					tokens[1].SourceInfo,
					fmt.Sprintf("invalid bound variable name: %s", tokens[1].Value),
				}
			}
			bound := &expr.Var{
				SI:   tokens[1].SourceInfo,
				Name: tokens[1].Value,
			}
			body, err := Expression(tokens[2:])
			if err != nil {
				return nil, err
			}
			if body == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no body in abstraction",
				}
			}
			e = wrapAppl(e, &expr.Abst{
				Bound: bound,
				Body:  body,
			})
			return e, nil

		case ";":
			afterSemicolon, err := Expression(tokens[1:])
			if err != nil {
				return nil, err
			}
			e = wrapAppl(e, afterSemicolon)
			return e, nil

		default:
			if IsSpecial(tokens[0].Value) {
				return nil, &Error{
					tokens[0].SourceInfo,
					fmt.Sprintf("unexpected symbol: %s", tokens[0].Value),
				}
			}
			e = wrapAppl(e, &expr.Var{
				SI:   tokens[0].SourceInfo,
				Name: tokens[0].Value,
			})
			tokens = tokens[1:]
		}
	}

	return e, nil
}

func wrapAppl(left, right expr.Expr) expr.Expr {
	if left == nil {
		return right
	}
	return &expr.Appl{
		Left:  left,
		Right: right,
	}
}

func findClosingParen(tokens []Token) int {
	round := 0  // ()
	square := 0 // []
	curly := 0  // {}
	for i := range tokens {
		switch tokens[i].Value {
		case "(":
			round++
		case ")":
			round--
		case "[":
			square++
		case "]":
			square--
		case "{":
			curly++
		case "}":
			curly--
		}
		if round < 0 || square < 0 || curly < 0 {
			return -1
		}
		if round == 0 && square == 0 && curly == 0 {
			return i
		}
	}
	return -1
}
