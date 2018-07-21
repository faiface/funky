package parse

import (
	"fmt"
	"math/big"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/parse/parseinfo"
)

type LiteralKind int

const (
	LiteralIdentifier LiteralKind = iota
	LiteralNumber
	LiteralChar
	LiteralString
)

func LiteralKindOf(s string) LiteralKind {
	r0, size := utf8.DecodeRuneInString(s)
	r1 := rune(0)
	if len(s) > size {
		r1, _ = utf8.DecodeRuneInString(s[size:])
	}
	switch {
	case unicode.IsDigit(r0) || ((r0 == '-' || r0 == '+') && unicode.IsDigit(r1)):
		return LiteralNumber
	case r0 == '\'':
		return LiteralChar
	case r0 == '"':
		return LiteralString
	default:
		return LiteralIdentifier
	}
}

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

	beforeColon, _, afterColon := FindNextSpecialOrBinding(tree, ":")
	if afterColon != nil {
		e, err := TreeToExpr(beforeColon)
		if err != nil {
			return nil, err
		}
		t, err := TreeToType(afterColon)
		if err != nil {
			return nil, err
		}
		return e.WithTypeInfo(t), nil
	}

	switch tree := tree.(type) {
	case *Literal:
		switch LiteralKindOf(tree.Value) {
		case LiteralIdentifier:
			return &expr.Var{SI: tree.SourceInfo(), Name: tree.Value}, nil
		case LiteralNumber:
			i := big.NewInt(0)
			_, ok := i.SetString(tree.Value, 10)
			if ok {
				return &expr.Int{SI: tree.SourceInfo(), Value: i}, nil
			}
			f, err := strconv.ParseFloat(tree.Value, 64)
			if err != nil {
				return nil, &Error{tree.SourceInfo(), err.Error()}
			}
			return &expr.Float{SI: tree.SourceInfo(), Value: f}, nil
		case LiteralChar:
			s, err := strconv.Unquote(tree.Value)
			if err != nil {
				return nil, &Error{tree.SourceInfo(), err.Error()}
			}
			r := []rune(s)[0] // has only one rune, no need to check
			return &expr.Char{SI: tree.SourceInfo(), Value: r}, nil
		case LiteralString:
			s, err := strconv.Unquote(tree.Value)
			if err != nil {
				return nil, &Error{tree.SourceInfo(), err.Error()}
			}
			// string literal syntactic sugar unfold
			runes := []rune(s)
			var stringExpr expr.Expr = &expr.Var{SI: tree.SourceInfo(), Name: "empty"}
			for i := len(runes) - 1; i >= 0; i-- {
				stringExpr = &expr.Appl{
					Left: &expr.Appl{
						Left:  &expr.Var{SI: tree.SourceInfo(), Name: "::"},
						Right: &expr.Char{SI: tree.SourceInfo(), Value: runes[i]},
					},
					Right: stringExpr,
				}
			}
			return stringExpr, nil
		}

	case *Paren:
		switch tree.Kind {
		case "(":
			if tree.Inside == nil {
				return nil, &Error{tree.SourceInfo(), "nothing inside parentheses"}
			}
			return TreeToExpr(tree.Inside)
		case "[":
			// list literal syntactic sugar undolf
			var elems []expr.Expr
			inside := tree.Inside
			for inside != nil {
				elemTree, _, after := FindNextSpecialOrBinding(inside, ",")
				inside = after
				elem, err := TreeToExpr(elemTree)
				if err != nil {
					return nil, err
				}
				elems = append(elems, elem)
			}
			var listExpr expr.Expr = &expr.Var{SI: tree.SI, Name: "empty"}
			for i := len(elems) - 1; i >= 0; i-- {
				listExpr = &expr.Appl{
					Left: &expr.Appl{
						Left:  &expr.Var{SI: elems[i].SourceInfo(), Name: "::"},
						Right: elems[i],
					},
					Right: listExpr,
				}
			}
			return listExpr, nil
		}
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Kind)}

	case *Special:
		switch tree.Kind {
		case ";":
			return TreeToExpr(tree.After)
		case "switch":
			expTree, caseBindingTree, nextCasesTree := FindNextSpecialOrBinding(tree.After, "case")
			if expTree == nil {
				return nil, &Error{tree.SourceInfo(), "no expression to switch"}
			}
			exp, err := TreeToExpr(expTree)
			if err != nil {
				return nil, err
			}
			sw := &expr.Switch{SI: tree.SourceInfo(), Expr: exp}
			for caseBindingTree != nil {
				caseBodyTree, newCaseBindingTree, newNextCasesTree := FindNextSpecialOrBinding(nextCasesTree, "case")

				caseBinding := caseBindingTree.(*Binding)
				altExpr, err := TreeToExpr(caseBinding.Bound)
				if err != nil {
					return nil, err
				}
				alt, ok := altExpr.(*expr.Var)
				if !ok {
					return nil, &Error{altExpr.SourceInfo(), "union alternative must be a simple variable"}
				}
				if alt.TypeInfo() != nil {
					return nil, &Error{altExpr.SourceInfo(), "union alternative name cannot have type"}
				}

				body, err := TreeToExpr(caseBodyTree)
				if err != nil {
					return nil, err
				}

				sw.Cases = append(sw.Cases, struct {
					SI   *parseinfo.Source
					Alt  string
					Body expr.Expr
				}{caseBindingTree.SourceInfo(), alt.Name, body})

				caseBindingTree = newCaseBindingTree
				nextCasesTree = newNextCasesTree
			}
			return sw, nil
		}
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Kind)}

	case *Binding:
		switch tree.Kind {
		case "\\":
			bound, err := TreeToExpr(tree.Bound)
			if err != nil {
				return nil, err
			}
			boundVar, ok := bound.(*expr.Var)
			if !ok {
				return nil, &Error{tree.SourceInfo(), "bound expression must be a simple variable"}
			}
			body, err := TreeToExpr(tree.After)
			if err != nil {
				return nil, err
			}
			return &expr.Abst{SI: tree.SourceInfo(), Bound: boundVar, Body: body}, nil
		}
		return nil, &Error{tree.SourceInfo(), fmt.Sprintf("unexpected: %s", tree.Kind)}

	case *Prefix:
		left, err := TreeToExpr(tree.Left)
		if err != nil {
			return nil, err
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
		SI:    si,
		Bound: &expr.Var{Name: "f"},
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
