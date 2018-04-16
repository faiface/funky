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
		if IsTypeName(tree.Value) {
			return &types.Appl{SI: tree.SI, Name: tree.Value}, nil
		}
		if IsTypeVar(tree.Value) || tree.Value == "->" {
			return &types.Var{SI: tree.SI, Name: tree.Value}, nil
		}
		return nil, &Error{tree.SI, fmt.Sprintf("invalid type identifier: %s", tree.Value)}

	case *Paren:
		switch tree.Type {
		case "(":
			return TreeToType(tree.Inside)
		}
		return nil, &Error{tree.SI, fmt.Sprintf("unexpected: %s", tree.Type)}

	case *Special:
		return nil, &Error{tree.SI, fmt.Sprintf("unexpected: %s", tree.Type)}

	case *Lambda:
		return nil, &Error{tree.SI, fmt.Sprintf("unexpected: %s", tree.Type)}

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
		inVar, ok := in.(*types.Var)
		if !ok || inVar.Name != "->" {
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
