package parse

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/faiface/funky/types"
)

func IsTypeName(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

func IsTypeVar(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsLower(r)
}

func Type(tokens []Token) (types.Type, error) {
	tree, err := MultiTree(tokens)
	if err != nil {
		return nil, err
	}
	return TreeToType(tree)
}

func TreeToType(tree Tree) (types.Type, error) {
	if tree == nil {
		return nil, nil
	}

	switch tree := tree.(type) {
	case *Literal:
		switch LiteralKindOf(tree.Value) {
		case LiteralIdentifier:
			// OK
		default:
			return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Value)}
		}
		if IsTypeName(tree.Value) || tree.Value == "->" {
			return &types.Appl{SI: tree.SourceInfo(), Name: tree.Value}, nil
		}
		if IsTypeVar(tree.Value) {
			return &types.Var{SI: tree.SourceInfo(), Name: tree.Value}, nil
		}
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("invalid type identifier: %s", tree.Value)}

	case *Paren:
		switch tree.Kind {
		case "(":
			if tree.Inside == nil {
				return nil, &Error{tree.SourceInfo(), "nothing inside parentheses"}
			}
			return TreeToType(tree.Inside)
		}
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Kind)}

	case *Special:
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Kind)}

	case *Binding:
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Kind)}

	case *Prefix:
		left, err := TreeToType(tree.Left)
		if err != nil {
			return nil, err
		}
		leftAppl, ok := left.(*types.Appl)
		if !ok {
			return nil, &Error{
				left.SourceInfo(),
				fmt.Sprintf("not a type constructor: %v", left),
			}
		}
		right, err := TreeToType(tree.Right)
		if err != nil {
			return nil, err
		}
		leftAppl.Args = append(leftAppl.Args, right)
		return leftAppl, nil

	case *Infix:
		in, err := TreeToType(tree.In)
		if err != nil {
			return nil, err
		}
		left, err := TreeToType(tree.Left)
		if err != nil {
			return nil, err
		}
		right, err := TreeToType(tree.Right)
		if err != nil {
			return nil, err
		}
		inAppl, ok := in.(*types.Appl)
		if !ok || inAppl.Name != "->" || len(inAppl.Args) != 0 {
			return nil, &Error{
				left.SourceInfo(),
				fmt.Sprintf("not a type constructor: %v", in),
			}
		}
		if left == nil || right == nil {
			return nil, &Error{
				in.SourceInfo(),
				"missing operands in function type",
			}
		}
		return &types.Func{
			From: left,
			To:   right,
		}, nil
	}

	panic("unreachable")
}
