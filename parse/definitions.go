package parse

import (
	"fmt"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/types"
)

func Definitions(tokens []Token) (names map[string]types.Name, defs map[string][]expr.Expr, err error) {
	tree, err := MultiTree(tokens)
	if err != nil {
		return nil, nil, err
	}
	return TreeToDefinitions(tree)
}

func TreeToDefinitions(tree Tree) (names map[string]types.Name, defs map[string][]expr.Expr, err error) {
	names = make(map[string]types.Name)
	defs = make(map[string][]expr.Expr)

	for tree != nil {
		before, at, after := FindNextSpecial(tree, "record", "enum", "alias", "def")
		if before != nil {
			return nil, nil, &Error{
				tree.SourceInfo(),
				fmt.Sprintf("expected record, enum, alias or def"),
			}
		}
		definition, next, _ := FindNextSpecial(after, "record", "enum", "alias", "def")
		tree = next
		switch at.(*Special).Type {
		case "def":
			name, e, err := treeToDef(definition)
			if err != nil {
				return nil, nil, err
			}
			defs[name] = append(defs[name], e)
		}
	}

	return names, defs, nil
}

func treeToDef(tree Tree) (string, expr.Expr, error) {
	nameTree, _, after := FindNextSpecial(tree, ":")
	typTree, _, bodyTree := FindNextSpecial(after, "=")

	var name string

	if lit, ok := nameTree.(*Literal); ok {
		name = lit.Value
	} else if infix, ok := nameTree.(*Infix); ok && infix.Left == nil && infix.Right == nil {
		if lit, ok := infix.In.(*Literal); ok {
			name = lit.Value
		}
	}

	if name == "" {
		return "", nil, &Error{
			nameTree.SourceInfo(),
			"definition name must be simple identifier",
		}
	}

	typ, err := TreeToType(typTree)
	if err != nil {
		return "", nil, err
	}
	body, err := TreeToExpr(bodyTree)
	if err != nil {
		return "", nil, err
	}

	if body.TypeInfo() != nil {
		return "", nil, &Error{
			bodyTree.SourceInfo(),
			"body type info must only be in definition",
		}
	}

	return name, body.WithTypeInfo(typ), nil
}
