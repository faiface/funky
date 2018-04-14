package parse

import (
	"github.com/faiface/funky/expr"
)

func Defs(tokens []Token) (map[string][]expr.Expr, error) {
	tree, err := MultiTree(tokens)
	if err != nil {
		return nil, err
	}
	return TreeToDefs(tree)
}

func TreeToDefs(tree Tree) (map[string][]expr.Expr, error) {
	defs := make(map[string][]expr.Expr)

	for tree != nil {
		before, _, after := FindNextSpecial(tree, "def")
		if before != nil {
			return nil, &Error{
				tree.SourceInfo(),
				"expected: def",
			}
		}
		nameTree, _, after := FindNextSpecial(after, ":")
		typTree, _, after := FindNextSpecial(after, "=")
		bodyTree, nextDef, _ := FindNextSpecial(after, "def")
		tree = nextDef

		nameLit, ok := nameTree.(*Literal)
		if !ok {
			return nil, &Error{
				nameTree.SourceInfo(),
				"definition name must be simple identifier",
			}
		}

		typ, err := TreeToType(typTree)
		if err != nil {
			return nil, err
		}
		body, err := TreeToExpr(bodyTree)
		if err != nil {
			return nil, err
		}

		if body.TypeInfo() != nil {
			return nil, &Error{
				bodyTree.SourceInfo(),
				"body type info must only be in definition",
			}
		}

		defs[nameLit.Value] = append(defs[nameLit.Value], body.WithTypeInfo(typ))
	}

	return defs, nil
}
