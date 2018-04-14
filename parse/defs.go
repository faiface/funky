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

/*func Definitions(tokens []Token) (map[string][]expr.Expr, error) {
	//TODO: redo
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

		e, err := Expr(tokens[equals+1 : nextDef])
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
}*/
