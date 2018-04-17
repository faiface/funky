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
		before, at, after := FindNextSpecial(tree, "record", "enum", "alias", "func")
		if before != nil {
			return nil, &Error{
				tree.SourceInfo(),
				fmt.Sprintf("expected record, enum, alias or func"),
			}
		}
		definition, next, _ := FindNextSpecial(after, "record", "enum", "alias", "func")
		tree = next

		switch at.(*Special).Type {
		case "record":
			name, record, err := treeToRecord(definition)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, Definition{name, record})

		case "func":
			name, body, err := treeToFunc(definition)
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
		return "", nil, &Error{nameLit.SourceInfo(), "invalid type name (must start with an upper-case letter)"}
	}
	var args []string
	for _, argTree := range header[1:] {
		argLit, ok := argTree.(*Literal)
		if !ok {
			return "", nil, &Error{argTree.SourceInfo(), "type argument must be a simple identifier"}
		}
		argName := argLit.Value
		if !IsTypeVar(argName) {
			return "", nil, &Error{argLit.SourceInfo(), "invalid type variable (must start with a lower-case letter)"}
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

func treeToFunc(tree Tree) (string, expr.Expr, error) {
	signatureTree, _, bodyTree := FindNextSpecial(tree, "=")

	if signatureTree == nil {
		return "", nil, &Error{tree.SourceInfo(), "missing function name"}
	}
	if bodyTree == nil {
		return "", nil, &Error{tree.SourceInfo(), "missing function body"}
	}

	sigExpr, err := TreeToExpr(signatureTree)
	if err != nil {
		return "", nil, err
	}
	signature, ok := sigExpr.(*expr.Var)
	if !ok {
		return "", nil, &Error{tree.SourceInfo(), "function name must be simple variable"}
	}

	if IsTypeName(signature.Name) {
		return "", nil, &Error{
			signature.SourceInfo(),
			"function name cannot start with an upper-case letter",
		}
	}
	if signature.TypeInfo() == nil {
		return "", nil, &Error{
			signature.SourceInfo(),
			"missing function type",
		}
	}

	body, err := TreeToExpr(bodyTree)
	if err != nil {
		return "", nil, err
	}

	if body.TypeInfo() != nil {
		return "", nil, &Error{
			bodyTree.SourceInfo(),
			"body type must only be in signature",
		}
	}

	return signature.Name, body.WithTypeInfo(signature.TypeInfo()), nil
}
