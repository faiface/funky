package parse

import (
	"fmt"

	"github.com/faiface/funky/expr"
)

func Expression(tokens []Token) (expr.Expr, error) {
	var e expr.Expr

	for len(tokens) > 0 {
		switch tokens[0].Value {
		case ")":
			return nil, &Error{
				tokens[0].SourceInfo,
				"no matching opening parenthesis",
			}

		case "(":
			closing, ok := findClosingParen(tokens)
			if !ok {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no matching closing parenthesis",
				}
			}
			inParen, err := Expression(tokens[1:closing])
			if err != nil {
				return nil, err
			}
			if inParen == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no expression in parentheses",
				}
			}
			e = wrapExprAppl(e, inParen)
			tokens = tokens[closing+1:]

		case "\\", "λ":
			if len(tokens) < 2 {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no bound variable in abstraction",
				}
			}
			bound, boundEnd, err := parseBound(tokens[1:])
			boundEnd++
			if err != nil {
				return nil, err
			}
			body, err := Expression(tokens[boundEnd:])
			if err != nil {
				return nil, err
			}
			if body == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no body in abstraction",
				}
			}
			e = wrapExprAppl(e, &expr.Abst{
				Bound: bound,
				Body:  body,
			})
			return e, nil

		case ";":
			afterSemicolon, err := Expression(tokens[1:])
			if err != nil {
				return nil, err
			}
			if afterSemicolon == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no expression after ;",
				}
			}
			e = wrapExprAppl(e, afterSemicolon)
			return e, nil

		case ":": // not a special symbol, must be space-separated
			if e == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no expression before :",
				}
			}
			t, err := Type(tokens[1:])
			if err != nil {
				return nil, err
			}
			if t == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no type after :",
				}
			}
			if e.TypeInfo() != nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"expression already has type info",
				}
			}
			e = e.WithTypeInfo(t)
			return e, nil

		default:
			if IsReserved(tokens[0].Value) {
				return nil, &Error{
					tokens[0].SourceInfo,
					fmt.Sprintf("unexpected: %s", tokens[0].Value),
				}
			}
			e = wrapExprAppl(e, &expr.Var{
				SI:   tokens[0].SourceInfo,
				Name: tokens[0].Value,
			})
			tokens = tokens[1:]
		}
	}

	return e, nil
}

func parseBound(tokens []Token) (*expr.Var, int, error) {
	var (
		bound expr.Expr
		end   int
		err   error
	)
	if tokens[0].Value == "(" {
		closing, ok := findClosingParen(tokens)
		if !ok {
			return nil, 0, &Error{
				tokens[0].SourceInfo,
				"no matching closing parenthesis",
			}
		}
		bound, err = Expression(tokens[:closing+1])
		end = closing + 1
	} else {
		bound, err = Expression(tokens[:1])
		end = 1
	}
	if err != nil {
		return nil, 0, err
	}
	boundVar, ok := bound.(*expr.Var)
	if !ok {
		return nil, 0, &Error{
			tokens[0].SourceInfo,
			"bound expression not a variable",
		}
	}
	return boundVar, end, nil
}

func wrapExprAppl(left, right expr.Expr) expr.Expr {
	if left == nil {
		return right
	}
	return &expr.Appl{
		Left:  left,
		Right: right,
	}
}

func findClosingParen(tokens []Token) (i int, ok bool) {
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
			return i, false
		}
		if round == 0 && square == 0 && curly == 0 {
			return i, true
		}
	}
	return len(tokens), false
}
