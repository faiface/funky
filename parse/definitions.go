package parse

import (
	"fmt"

	"github.com/faiface/funky/expr"
)

func Definitions(tokens []Token) (map[string][]expr.Expr, error) {
	defs := make(map[string][]expr.Expr)

	for len(tokens) > 0 {
		name := tokens[0].Value
		if IsReserved(name) {
			return nil, &Error{
				tokens[0].SourceInfo,
				fmt.Sprintf("unexpected symbol: %s", tokens[0].Value),
			}
		}
		if len(tokens) < 2 || tokens[1].Value != ":" {
			return nil, &Error{
				tokens[0].SourceInfo,
				"expected :",
			}
		}
		equals := findNextTopLevel("=", tokens)
		if equals == -1 {
			return nil, &Error{
				tokens[len(tokens)-1].SourceInfo,
				"expected =",
			}
		}
		typ, err := Type(tokens[2:equals])
		if err != nil {
			return nil, err
		}
		if typ == nil {
			return nil, &Error{
				tokens[1].SourceInfo,
				"no type information after :",
			}
		}
		nextColon := findNextTopLevel(":", tokens[equals:])
		nextDef := nextColon - 1 + equals
		if nextColon == -1 {
			nextDef = len(tokens)
		}
		exp, err := Expression(tokens[equals+1 : nextDef])
		if err != nil {
			return nil, err
		}
		if exp == nil {
			return nil, &Error{
				tokens[equals].SourceInfo,
				"no expression after =",
			}
		}
		exp = exp.WithTypeInfo(typ)
		defs[name] = append(defs[name], exp)
		tokens = tokens[nextDef:]
	}

	return defs, nil
}

func findNextTopLevel(symbol string, tokens []Token) int {
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
		if tokens[i].Value == symbol && round == 0 && square == 0 && curly == 0 {
			return i
		}
	}
	return -1
}
