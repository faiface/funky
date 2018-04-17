package parse

import (
	"fmt"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/types"
)

type Definition struct {
	Name  string
	Value interface{} // expr.Expr, *types.Record, *types.Enum, *types.Alias
}

func Definitions(tokens []Token) ([]Definition, error) {
	tree, err := MultiTree(tokens)
	if err != nil {
		return nil, err
	}
	return TreeToDefinitions(tree)
}

func TreeToDefinitions(tree Tree) ([]Definition, error) {
	var definitions []Definition

	for tree != nil {
		before, at, after := FindNextSpecial(tree, "record", "enum", "alias", "def")
		if before != nil {
			return nil, &Error{
				tree.SourceInfo(),
				fmt.Sprintf("expected record, enum, alias or def"),
			}
		}
		definition, next, _ := FindNextSpecial(after, "record", "enum", "alias", "def")
		tree = next

		switch at.(*Special).Type {
		case "record":
			name, record, err := treeToRecord(definition)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, Definition{name, record})

		case "def":
			name, body, err := treeToDef(definition)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, Definition{name, body})
		}
	}

	return definitions, nil
}

func treeToRecord(tree Tree) (name string, record *types.Record, err error) {
	headerTree, _, fieldsTree := FindNextSpecial(tree, "=")

	header := Flatten(headerTree)
	if len(header) == 0 {
		return "", nil, &Error{tree.SourceInfo(), "missing type name"}
	}
	nameLit, ok := header[0].(*Literal)
	if !ok {
		return "", nil, &Error{header[0].SourceInfo(), "type name must be a simple identifier"}
	}
	name = nameLit.Value
	if !IsTypeName(name) {
		return "", nil, &Error{nameLit.SI, "invalid type name (must start with an upper-case letter)"}
	}
	var args []string
	for _, argTree := range header[1:] {
		argLit, ok := argTree.(*Literal)
		if !ok {
			return "", nil, &Error{argTree.SourceInfo(), "type argument must be a simple identifier"}
		}
		argName := argLit.Value
		if !IsTypeVar(argName) {
			return "", nil, &Error{argLit.SI, "invalid type variable (must start with a lower-case letter)"}
		}
		args = append(args, argName)
	}

	var fields []types.Field

	for fieldsTree != nil {
		fieldTree, _, after := FindNextSpecial(fieldsTree, ",")
		fieldsTree = after

		if fieldTree == nil {
			continue
		}

		field, err := TreeToExpr(fieldTree)
		if err != nil {
			return "", nil, err
		}
		fieldVar, ok := field.(*expr.Var)
		if !ok {
			return "", nil, &Error{field.SourceInfo(), "record field must be a simple variable"}
		}
		if fieldVar.TypeInfo() == nil {
			return "", nil, &Error{field.SourceInfo(), "missing record field type"}
		}

		fields = append(fields, types.Field{
			SI:   fieldVar.SourceInfo(),
			Name: fieldVar.Name,
			Type: fieldVar.TypeInfo(),
		})
	}

	return name, &types.Record{
		SI:     tree.SourceInfo(),
		Args:   args,
		Fields: fields,
	}, nil
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
	if IsTypeName(name) {
		return "", nil, &Error{
			nameTree.SourceInfo(),
			"definition name cannot start with an upper-case letter",
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
