package parse

import (
	"fmt"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/parse/parseinfo"
)

func Expr(tokens []Token) (expr.Expr, error) {
	tree, err := MultiTree(tokens)
	if err != nil {
		return nil, err
	}
	return TreeToExpr(tree)
}

func TreeToExpr(tree Tree) (expr.Expr, error) {
	if tree == nil {
		return nil, nil
	}

	switch tree := tree.(type) {
	case *Literal:
		return &expr.Var{SI: tree.SourceInfo(), Name: tree.Value}, nil

	case *Paren:
		switch tree.Type {
		case "(":
			return TreeToExpr(tree.Inside)
		}
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Type)}

	case *Special:
		switch tree.Type {
		case ";":
			return TreeToExpr(tree.After)
		}
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Type)}

	case *Lambda:
		bound, err := TreeToExpr(tree.Bound)
		if err != nil {
			return nil, err
		}
		boundVar, ok := bound.(*expr.Var)
		if !ok {
			return nil, &Error{tree.SourceInfo(), "bound expression not a variable"}
		}
		body, err := TreeToExpr(tree.After)
		if err != nil {
			return nil, err
		}
		return &expr.Abst{Bound: boundVar, Body: body}, nil

	case *Prefix:
		left, err := TreeToExpr(tree.Left)
		if err != nil {
			return nil, err
		}
		if special, ok := tree.Right.(*Special); ok && special.Type == ":" { // type info after :
			typ, err := TreeToType(special.After)
			if err != nil {
				return nil, err
			}
			return left.WithTypeInfo(typ), nil
		}
		right, err := TreeToExpr(tree.Right)
		if err != nil {
			return nil, err
		}
		if right == nil {
			return left, nil
		}
		return &expr.Appl{Left: left, Right: right}, nil

	case *Infix:
		in, err := TreeToExpr(tree.In)
		if err != nil {
			return nil, err
		}
		left, err := TreeToExpr(tree.Left)
		if err != nil {
			return nil, err
		}
		right, err := TreeToExpr(tree.Right)
		if err != nil {
			return nil, err
		}
		switch {
		case left == nil && right == nil: // (+)
			return in, nil
		case right == nil: // (1 +)
			return &expr.Appl{Left: in, Right: left}, nil
		case left == nil: // (+ 2)
			return &expr.Appl{
				Left:  &expr.Appl{Left: newFlipExpr(in.SourceInfo()), Right: in},
				Right: right,
			}, nil
		default: // (1 + 2)
			return &expr.Appl{
				Left:  &expr.Appl{Left: in, Right: left},
				Right: right,
			}, nil
		}
	}

	panic("unreachable")
}

func newFlipExpr(si *parseinfo.Source) expr.Expr {
	return &expr.Abst{
		Bound: &expr.Var{SI: si, Name: "f"},
		Body: &expr.Abst{
			Bound: &expr.Var{Name: "x"},
			Body: &expr.Abst{
				Bound: &expr.Var{Name: "y"},
				Body: &expr.Appl{
					Left: &expr.Appl{
						Left:  &expr.Var{Name: "f"},
						Right: &expr.Var{Name: "y"},
					},
					Right: &expr.Var{Name: "x"},
				},
			},
		},
	}
}
