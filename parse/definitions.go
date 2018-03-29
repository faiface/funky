package parse

import (
	"fmt"

	"github.com/faiface/funky/expr"
)

func Definitions(tokens []Token) (map[string][]expr.Expr, error) {
	defs := make(map[string][]expr.Expr)

	for len(tokens) > 0 {
		if len(tokens) < 4 {
			return nil, &Error{
				tokens[len(tokens)-1].SourceInfo,
				"incomplete definition",
			}
		}
		if tokens[0].Value != "def" {
			return nil, &Error{
				tokens[0].SourceInfo,
				"expected: def",
			}
		}

		name := tokens[1].Value
		if IsReserved(name) {
			return nil, &Error{
				tokens[1].SourceInfo,
				fmt.Sprintf("unexpected symbol: %s", tokens[0].Value),
			}
		}

		if tokens[2].Value != ":" {
			return nil, &Error{
				tokens[2].SourceInfo,
				"expected: :",
			}
		}
		equals := 3
		for equals < len(tokens) && tokens[equals].Value != "=" {
			equals++
		}
		if equals == len(tokens) {
			return nil, &Error{
				tokens[len(tokens)-1].SourceInfo,
				"expected: =",
			}
		}

		t, err := Type(tokens[3:equals])
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, &Error{
				tokens[2].SourceInfo,
				"no type after :",
			}
		}

		nextDef := equals + 1
		for nextDef < len(tokens) && tokens[nextDef].Value != "def" {
			nextDef++
		}

		e, err := Expression(tokens[equals+1 : nextDef])
		if err != nil {
			return nil, err
		}
		if e == nil {
			return nil, &Error{
				tokens[equals].SourceInfo,
				"no expression after =",
			}
		}

		e = e.WithTypeInfo(t)
		defs[name] = append(defs[name], e)

		tokens = tokens[nextDef:]
	}

	return defs, nil
}
