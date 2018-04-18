package parse

import (
	"fmt"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/types"
)

type Definition struct {
	Name  string
	Value interface{} // expr.Expr, *types.Record, *types.Union, *types.Alias
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
		before, at, after := FindNextSpecial(tree, "record", "union", "alias", "func")
		if before != nil {
			return nil, &Error{
				tree.SourceInfo(),
				fmt.Sprintf("expected record, union, alias or func"),
			}
		}
		definition, next, _ := FindNextSpecial(after, "record", "union", "alias", "func")
		tree = next

		switch at.(*Special).Kind {
		case "record":
			name, record, err := treeToRecord(definition)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, Definition{name, record})

		case "union":
			name, union, err := treeToUnion(definition)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, Definition{name, union})

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

func treeToTypeHeader(tree Tree) (name string, args []string, err error) {
	header := Flatten(tree)
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
	return name, args, nil
}

func treeToRecord(tree Tree) (name string, record *types.Record, err error) {
	headerTree, _, fieldsTree := FindNextSpecial(tree, "=")

	name, args, err := treeToTypeHeader(headerTree)
	if err != nil {
		return "", nil, err
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
			return "", nil, &Error{field.SourceInfo(), "record field must be simple variable"}
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

func treeToUnion(tree Tree) (name string, union *types.Union, err error) {
	headerTree, _, altsTree := FindNextSpecial(tree, "=")

	name, args, err := treeToTypeHeader(headerTree)
	if err != nil {
		return "", nil, err
	}

	var alts []types.Alternative

	for altsTree != nil {
		altTree, _, after := FindNextSpecial(altsTree, "|")
		altsTree = after

		if altTree == nil {
			continue
		}

		fieldTrees := Flatten(altTree)

		altNameExpr, err := TreeToExpr(fieldTrees[0])
		if err != nil {
			return "", nil, err
		}
		if altNameExpr.TypeInfo() != nil {
			return "", nil, &Error{altNameExpr.SourceInfo(), "union alternative name cannot have type"}
		}
		altNameVar, ok := altNameExpr.(*expr.Var)
		if !ok {
			return "", nil, &Error{altNameExpr.SourceInfo(), "union alternative name must be simple variable"}
		}
		altName := altNameVar.Name

		var fields []types.Type
		for _, fieldTree := range fieldTrees[1:] {
			field, err := TreeToType(fieldTree)
			if err != nil {
				return "", nil, err
			}
			fields = append(fields, field)
		}

		alts = append(alts, types.Alternative{
			SI:     altTree.SourceInfo(),
			Name:   altName,
			Fields: fields,
		})
	}

	return name, &types.Union{
		SI:   tree.SourceInfo(),
		Args: args,
		Alts: alts,
	}, nil
}

func treeToFunc(tree Tree) (name string, body expr.Expr, err error) {
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

	body, err = TreeToExpr(bodyTree)
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
