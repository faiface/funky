package compile

import (
	"fmt"

	"github.com/faiface/crux"
	"github.com/faiface/crux/mk"
	"github.com/faiface/crux/runtime"
	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/parse"
	"github.com/faiface/funky/parse/parseinfo"
	"github.com/faiface/funky/types"
)

type Error struct {
	SourceInfo *parseinfo.Source
	Msg        string
}

func (err *Error) Error() string {
	return fmt.Sprintf("%v: %s", err.SourceInfo, err.Msg)
}

type Env struct {
	inited bool
	names  map[string]types.Name
	funcs  map[string][]funcImpl
}

type funcImpl interface {
	SourceInfo() *parseinfo.Source
	TypeInfo() types.Type
}

type (
	internal struct {
		SI   *parseinfo.Source
		Type types.Type
		Expr crux.Expr
	}

	function struct {
		Expr expr.Expr
	}
)

func (i *internal) SourceInfo() *parseinfo.Source { return i.SI }
func (f *function) SourceInfo() *parseinfo.Source { return f.Expr.SourceInfo() }
func (i *internal) TypeInfo() types.Type          { return i.Type }
func (f *function) TypeInfo() types.Type          { return f.Expr.TypeInfo() }

func parseType(s string) types.Type {
	tokens, err := parse.Tokenize("", s)
	if err != nil {
		panic(err)
	}
	typ, err := parse.Type(tokens)
	if err != nil {
		panic(err)
	}
	return typ
}

func (env *Env) lazyInit() {
	if env.inited {
		return
	}
	env.inited = true

	// built-in types
	env.names = map[string]types.Name{
		"Char":  &types.Builtin{NumArgs: 0},
		"Int":   &types.Builtin{NumArgs: 0},
		"Float": &types.Builtin{NumArgs: 0},
	}

	env.funcs = make(map[string][]funcImpl)

	// built-in operator functions

	// Char
	env.addFunc("int", &internal{Type: parseType("Char -> Int"), Expr: mk.Op(runtime.OpCharInt)})
	env.addFunc("inc", &internal{Type: parseType("Char -> Char"), Expr: mk.Op(runtime.OpCharInc)})
	env.addFunc("dec", &internal{Type: parseType("Char -> Char"), Expr: mk.Op(runtime.OpCharDec)})
	env.addFunc("+", &internal{Type: parseType("Char -> Int -> Char"), Expr: mk.Op(runtime.OpCharAdd)})
	env.addFunc("-", &internal{Type: parseType("Char -> Int -> Char"), Expr: mk.Op(runtime.OpCharSub)})
	env.addFunc("==", &internal{Type: parseType("Char -> Char -> Bool"), Expr: mk.Op(runtime.OpCharEq)})
	env.addFunc("!=", &internal{Type: parseType("Char -> Char -> Bool"), Expr: mk.Op(runtime.OpCharNeq)})
	env.addFunc("<", &internal{Type: parseType("Char -> Char -> Bool"), Expr: mk.Op(runtime.OpCharLess)})
	env.addFunc("<=", &internal{Type: parseType("Char -> Char -> Bool"), Expr: mk.Op(runtime.OpCharLessEq)})
	env.addFunc(">", &internal{Type: parseType("Char -> Char -> Bool"), Expr: mk.Op(runtime.OpCharMore)})
	env.addFunc(">=", &internal{Type: parseType("Char -> Char -> Bool"), Expr: mk.Op(runtime.OpCharMoreEq)})

	// Int
	env.addFunc("char", &internal{Type: parseType("Int -> Char"), Expr: mk.Op(runtime.OpIntChar)})
	env.addFunc("float", &internal{Type: parseType("Int -> Float"), Expr: mk.Op(runtime.OpIntFloat)})
	env.addFunc("string", &internal{Type: parseType("Int -> String"), Expr: mk.Op(runtime.OpIntString)})
	env.addFunc("neg", &internal{Type: parseType("Int -> Int"), Expr: mk.Op(runtime.OpIntNeg)})
	env.addFunc("abs", &internal{Type: parseType("Int -> Int"), Expr: mk.Op(runtime.OpIntAbs)})
	env.addFunc("inc", &internal{Type: parseType("Int -> Int"), Expr: mk.Op(runtime.OpIntInc)})
	env.addFunc("dec", &internal{Type: parseType("Int -> Int"), Expr: mk.Op(runtime.OpIntDec)})
	env.addFunc("+", &internal{Type: parseType("Int -> Int -> Int"), Expr: mk.Op(runtime.OpIntAdd)})
	env.addFunc("-", &internal{Type: parseType("Int -> Int -> Int"), Expr: mk.Op(runtime.OpIntSub)})
	env.addFunc("*", &internal{Type: parseType("Int -> Int -> Int"), Expr: mk.Op(runtime.OpIntMul)})
	env.addFunc("/", &internal{Type: parseType("Int -> Int -> Int"), Expr: mk.Op(runtime.OpIntDiv)})
	env.addFunc("%", &internal{Type: parseType("Int -> Int -> Int"), Expr: mk.Op(runtime.OpIntMod)})
	env.addFunc("^", &internal{Type: parseType("Int -> Int -> Int"), Expr: mk.Op(runtime.OpIntExp)})
	env.addFunc("==", &internal{Type: parseType("Int -> Int -> Bool"), Expr: mk.Op(runtime.OpIntEq)})
	env.addFunc("!=", &internal{Type: parseType("Int -> Int -> Bool"), Expr: mk.Op(runtime.OpIntNeq)})
	env.addFunc("<", &internal{Type: parseType("Int -> Int -> Bool"), Expr: mk.Op(runtime.OpIntLess)})
	env.addFunc("<=", &internal{Type: parseType("Int -> Int -> Bool"), Expr: mk.Op(runtime.OpIntLessEq)})
	env.addFunc(">", &internal{Type: parseType("Int -> Int -> Bool"), Expr: mk.Op(runtime.OpIntMore)})
	env.addFunc(">=", &internal{Type: parseType("Int -> Int -> Bool"), Expr: mk.Op(runtime.OpIntMoreEq)})

	// Float
	env.addFunc("int", &internal{Type: parseType("Float -> Int"), Expr: mk.Op(runtime.OpFloatInt)})
	env.addFunc("string", &internal{Type: parseType("Float -> String"), Expr: mk.Op(runtime.OpFloatString)})
	env.addFunc("neg", &internal{Type: parseType("Float -> Float"), Expr: mk.Op(runtime.OpFloatNeg)})
	env.addFunc("abs", &internal{Type: parseType("Float -> Float"), Expr: mk.Op(runtime.OpFloatAbs)})
	env.addFunc("inc", &internal{Type: parseType("Float -> Float"), Expr: mk.Op(runtime.OpFloatInc)})
	env.addFunc("dec", &internal{Type: parseType("Float -> Float"), Expr: mk.Op(runtime.OpFloatDec)})
	env.addFunc("+", &internal{Type: parseType("Float -> Float -> Float"), Expr: mk.Op(runtime.OpFloatAdd)})
	env.addFunc("-", &internal{Type: parseType("Float -> Float -> Float"), Expr: mk.Op(runtime.OpFloatSub)})
	env.addFunc("*", &internal{Type: parseType("Float -> Float -> Float"), Expr: mk.Op(runtime.OpFloatMul)})
	env.addFunc("/", &internal{Type: parseType("Float -> Float -> Float"), Expr: mk.Op(runtime.OpFloatDiv)})
	env.addFunc("%", &internal{Type: parseType("Float -> Float -> Float"), Expr: mk.Op(runtime.OpFloatMod)})
	env.addFunc("^", &internal{Type: parseType("Float -> Float -> Float"), Expr: mk.Op(runtime.OpFloatExp)})
	env.addFunc("==", &internal{Type: parseType("Float -> Float -> Bool"), Expr: mk.Op(runtime.OpFloatEq)})
	env.addFunc("!=", &internal{Type: parseType("Float -> Float -> Bool"), Expr: mk.Op(runtime.OpFloatNeq)})
	env.addFunc("<", &internal{Type: parseType("Float -> Float -> Bool"), Expr: mk.Op(runtime.OpFloatLess)})
	env.addFunc("<=", &internal{Type: parseType("Float -> Float -> Bool"), Expr: mk.Op(runtime.OpFloatLessEq)})
	env.addFunc(">", &internal{Type: parseType("Float -> Float -> Bool"), Expr: mk.Op(runtime.OpFloatMore)})
	env.addFunc(">=", &internal{Type: parseType("Float -> Float -> Bool"), Expr: mk.Op(runtime.OpFloatMoreEq)})

	// String
	env.addFunc("int", &internal{Type: parseType("String -> Int"), Expr: mk.Op(runtime.OpStringInt)})
	env.addFunc("float", &internal{Type: parseType("String -> Float"), Expr: mk.Op(runtime.OpStringFloat)})
}

func (env *Env) Add(d parse.Definition) error {
	env.lazyInit()

	switch value := d.Value.(type) {
	case *types.Record:
		return env.addRecord(d.Name, value)
	case *types.Union:
		return env.addUnion(d.Name, value)
	case *types.Alias:
		return env.addAlias(d.Name, value)
	case expr.Expr:
		return env.addFunc(d.Name, &function{value})
	}

	panic("unreachable")
}

func (env *Env) addRecord(name string, record *types.Record) error {
	if env.names[name] != nil {
		return &Error{
			record.SourceInfo(),
			fmt.Sprintf("type name %s already defined: %v", name, env.names[name].SourceInfo()),
		}
	}
	env.names[name] = record

	var args []types.Type
	for _, arg := range record.Args {
		args = append(args, &types.Var{Name: arg})
	}
	recordType := &types.Appl{
		SI:   record.SourceInfo(),
		Name: name,
		Args: args,
	}

	// add record constructor
	var constructorType types.Type = recordType
	for i := len(record.Fields) - 1; i >= 0; i-- {
		constructorType = &types.Func{
			From: record.Fields[i].Type,
			To:   constructorType,
		}
	}
	err := env.addFunc(
		name,
		&internal{
			SI:   record.SourceInfo(),
			Type: constructorType,
			Expr: mk.Make(0),
		},
	)
	if err != nil {
		return err
	}

	// add record field getters
	// RecordType -> FieldType
	for i, field := range record.Fields {
		err := env.addFunc(field.Name, &internal{
			SI:   field.SI,
			Type: &types.Func{From: recordType, To: field.Type},
			Expr: mk.Field(int32(i)),
		})
		if err != nil {
			return err
		}
	}

	// add record field setters
	// (FieldType -> FieldType) -> RecordType -> RecordType
	fieldVars := make([]string, len(record.Fields))
	for i := range fieldVars {
		fieldVars[i] = fmt.Sprintf("x%d", i)
	}
	switchArgs := append([]string{"f"}, fieldVars...)
	for i, field := range record.Fields {
		switchResult := mk.Appl(mk.Make(0), make([]crux.Expr, len(fieldVars))...)
		for j := range switchResult.Rands {
			switchResult.Rands[j] = mk.Var(fieldVars[j], -1)
		}
		switchResult.Rands[i] = mk.Appl(mk.Var("f", -1), mk.Var(fieldVars[i], -1))
		err := env.addFunc(field.Name, &internal{
			SI: field.SI,
			Type: &types.Func{
				From: &types.Func{From: field.Type, To: field.Type},
				To:   &types.Func{From: recordType, To: recordType},
			},
			Expr: mk.Abst("f", "r")(mk.Switch(mk.Var("r", -1),
				mk.Appl(mk.Abst(switchArgs...)(switchResult), mk.Var("f", -1)),
			)),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (env *Env) addUnion(name string, union *types.Union) error {
	if env.names[name] != nil {
		return &Error{
			union.SourceInfo(),
			fmt.Sprintf("type name %s already defined: %v", name, env.names[name].SourceInfo()),
		}
	}
	env.names[name] = union

	var args []types.Type
	for _, arg := range union.Args {
		args = append(args, &types.Var{Name: arg})
	}
	unionType := &types.Appl{
		SI:   union.SourceInfo(),
		Name: name,
		Args: args,
	}

	// add union alternative constructors
	for i, alt := range union.Alts {
		alternative := int32(i)
		var altType types.Type = unionType
		for i := len(alt.Fields) - 1; i >= 0; i-- {
			altType = &types.Func{
				From: alt.Fields[i],
				To:   altType,
			}
		}
		err := env.addFunc(
			alt.Name,
			&internal{
				SI:   alt.SI,
				Type: altType,
				Expr: mk.Make(alternative),
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (env *Env) addAlias(name string, alias *types.Alias) error {
	if env.names[name] != nil {
		return &Error{
			alias.SourceInfo(),
			fmt.Sprintf("type name %s already defined: %v", name, env.names[name].SourceInfo()),
		}
	}
	env.names[name] = alias
	return nil
}

func (env *Env) addFunc(name string, imp funcImpl) error {
	env.funcs[name] = append(env.funcs[name], imp)
	return nil
}
