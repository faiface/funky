package parse

import (
	"fmt"

	"github.com/faiface/funky/types"
)

func Type(tokens []Token) (types.Type, error) {
	var t types.Type

	for len(tokens) > 0 {
		switch tokens[0].Value {
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
			inParen, err := Type(tokens[1:closing])
			if err != nil {
				return nil, err
			}
			if inParen == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no type in parentheses",
				}
			}
			t, err = wrapTypeAppl(t, inParen)
			if err != nil {
				return nil, err
			}
			tokens = tokens[closing+1:]

		case "->":
			if t == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no left side in function type",
				}
			}
			rightSide, err := Type(tokens[1:])
			if err != nil {
				return nil, err
			}
			if rightSide == nil {
				return nil, &Error{
					tokens[0].SourceInfo,
					"no right side in function type",
				}
			}
			return &types.Func{From: t, To: rightSide}, nil

		default:
			if IsSpecial(tokens[0].Value) {
				return nil, &Error{
					tokens[0].SourceInfo,
					fmt.Sprintf("unexpected symbol: %s", tokens[0].Value),
				}
			}
			var ident types.Type
			if IsConsName(tokens[0].Value) {
				ident = &types.Appl{Cons: &types.Var{
					SI:   tokens[0].SourceInfo,
					Name: tokens[0].Value,
				}}
			} else {
				ident = &types.Var{
					SI:   tokens[0].SourceInfo,
					Name: tokens[0].Value,
				}
			}
			var err error
			t, err = wrapTypeAppl(t, ident)
			if err != nil {
				return nil, err
			}
			tokens = tokens[1:]
		}
	}

	return t, nil
}

func wrapTypeAppl(left, right types.Type) (types.Type, error) {
	switch left := left.(type) {
	case nil:
		return right, nil

	case *types.Appl:
		left.Args = append(left.Args, right)
		return left, nil

	default:
		return nil, &Error{
			right.SourceInfo(),
			fmt.Sprintf("not a type constructor: %v", left),
		}
	}
}
