package compile

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/parse"
	"github.com/faiface/funky/parse/parseinfo"
	"github.com/faiface/funky/runtime"
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
	internalFunc struct {
		SI     *parseinfo.Source
		Type   types.Type
		Arity  int
		GoFunc runtime.GoFunc
	}

	exprFunc struct {
		Expr expr.Expr
	}
)

func (i *internalFunc) SourceInfo() *parseinfo.Source { return i.SI }
func (e *exprFunc) SourceInfo() *parseinfo.Source     { return e.Expr.SourceInfo() }
func (i *internalFunc) TypeInfo() types.Type          { return i.Type }
func (e *exprFunc) TypeInfo() types.Type              { return e.Expr.TypeInfo() }

func runtimeValueToString(s runtime.Value) string {
	var builder strings.Builder
	union := s.Reduce().(*runtime.Union)
	for union.Alternative != 0 {
		builder.WriteRune(union.Fields[0].Reduce().(*runtime.Char).Value)
		union = union.Fields[1].Reduce().(*runtime.Union)
	}
	return builder.String()
}

func boolToRuntimeValue(b bool) runtime.Value {
	alt := 0
	if !b {
		alt = 1
	}
	return &runtime.Union{Alternative: alt, Fields: nil}
}

func stringToRuntimeValue(s string) runtime.Value {
	str := &runtime.Union{Alternative: 0}
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		str = &runtime.Union{Alternative: 1, Fields: []runtime.Value{&runtime.Char{Value: runes[i]}, str}}
	}
	return str
}

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

	// built-in functions

	// common
	env.addFunc("eval", &internalFunc{
		Type: parseType("a -> b -> b"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			d.Next.Value.Reduce()
			return d.Value.Reduce()
		},
	})
	env.addFunc("dump", &internalFunc{
		Type: parseType("String -> a -> a"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x, y := d.Next.Value, d.Value
			fmt.Fprint(os.Stderr, runtimeValueToString(x))
			return y
		},
	})
	env.addFunc("error", &internalFunc{
		Type: parseType("String -> a"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", runtimeValueToString(x))
			os.Exit(1)
			return nil
		},
	})

	// conversions
	env.addFunc("int", &internalFunc{
		Type: parseType("Char -> Int"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value
			return &runtime.Int{Value: big.NewInt(int64(x.Reduce().(*runtime.Char).Value))}
		},
	})
	env.addFunc("char", &internalFunc{
		Type: parseType("Int -> Char"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value
			return &runtime.Char{Value: rune(x.Reduce().(*runtime.Int).Value.Int64())}
		},
	})
	env.addFunc("int", &internalFunc{
		Type: parseType("Float -> Int"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := math.Floor(d.Value.Reduce().(*runtime.Float).Value)
			z, _ := big.NewFloat(x).Int(nil)
			return &runtime.Int{Value: z}
		},
	})
	env.addFunc("float", &internalFunc{
		Type: parseType("Int -> Float"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value.Reduce().(*runtime.Int).Value
			z, _ := big.NewFloat(0).SetInt(x).Float64()
			return &runtime.Float{Value: z}
		},
	})
	env.addFunc("string", &internalFunc{
		Type: parseType("Int -> String"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value.Reduce().(*runtime.Int).Value
			return stringToRuntimeValue(x.Text(10))
		},
	})
	env.addFunc("int", &internalFunc{
		Type: parseType("String -> Maybe Int"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			s := runtimeValueToString(d.Value)
			i, ok := big.NewInt(0).SetString(s, 10)
			if !ok {
				return &runtime.Union{Alternative: 0}
			}
			return &runtime.Union{Alternative: 1, Fields: []runtime.Value{&runtime.Int{Value: i}}}
		},
	})
	env.addFunc("string", &internalFunc{
		Type: parseType("Float -> String"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value.Reduce().(*runtime.Float).Value
			return stringToRuntimeValue(strconv.FormatFloat(x, 'f', -1, 64))
		},
	})
	env.addFunc("float", &internalFunc{
		Type: parseType("String -> Maybe Float"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			s := runtimeValueToString(d.Value)
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return &runtime.Union{Alternative: 0}
			}
			return &runtime.Union{Alternative: 1, Fields: []runtime.Value{&runtime.Float{Value: f}}}
		},
	})

	// Char
	env.addFunc("==", &internalFunc{
		Type: parseType("Char -> Char -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Char).Value
			y := d.Value.Reduce().(*runtime.Char).Value
			return boolToRuntimeValue(x == y)
		},
	})
	env.addFunc("!=", &internalFunc{
		Type: parseType("Char -> Char -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Char).Value
			y := d.Value.Reduce().(*runtime.Char).Value
			return boolToRuntimeValue(x != y)
		},
	})

	// Int
	env.addFunc("neg", &internalFunc{
		Type: parseType("Int -> Int"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value.Reduce().(*runtime.Int).Value
			z := big.NewInt(0)
			z.Neg(x)
			return &runtime.Int{Value: z}
		},
	})
	env.addFunc("+", &internalFunc{
		Type: parseType("Int -> Int -> Int"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			z := big.NewInt(0)
			z.Add(x, y)
			return &runtime.Int{Value: z}
		},
	})
	env.addFunc("-", &internalFunc{
		Type: parseType("Int -> Int -> Int"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			z := big.NewInt(0)
			z.Sub(x, y)
			return &runtime.Int{Value: z}
		},
	})
	env.addFunc("*", &internalFunc{
		Type: parseType("Int -> Int -> Int"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			z := big.NewInt(0)
			z.Mul(x, y)
			return &runtime.Int{Value: z}
		},
	})
	env.addFunc("/", &internalFunc{
		Type: parseType("Int -> Int -> Int"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			z := big.NewInt(0)
			z.Div(x, y)
			return &runtime.Int{Value: z}
		},
	})
	env.addFunc("%", &internalFunc{
		Type: parseType("Int -> Int -> Int"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			z := big.NewInt(0)
			z.Mod(x, y)
			return &runtime.Int{Value: z}
		},
	})
	env.addFunc("^", &internalFunc{
		Type: parseType("Int -> Int -> Int"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			z := big.NewInt(0)
			z.Exp(x, y, nil)
			return &runtime.Int{Value: z}
		},
	})
	env.addFunc("==", &internalFunc{
		Type: parseType("Int -> Int -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			return boolToRuntimeValue(x.Cmp(y) == 0)
		},
	})
	env.addFunc("!=", &internalFunc{
		Type: parseType("Int -> Int -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			return boolToRuntimeValue(x.Cmp(y) != 0)
		},
	})
	env.addFunc("<", &internalFunc{
		Type: parseType("Int -> Int -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			return boolToRuntimeValue(x.Cmp(y) < 0)
		},
	})
	env.addFunc("<=", &internalFunc{
		Type: parseType("Int -> Int -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			return boolToRuntimeValue(x.Cmp(y) <= 0)
		},
	})
	env.addFunc(">", &internalFunc{
		Type: parseType("Int -> Int -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			return boolToRuntimeValue(x.Cmp(y) > 0)
		},
	})
	env.addFunc(">=", &internalFunc{
		Type: parseType("Int -> Int -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Int).Value
			y := d.Value.Reduce().(*runtime.Int).Value
			return boolToRuntimeValue(x.Cmp(y) >= 0)
		},
	})

	// Float
	env.addFunc("neg", &internalFunc{
		Type: parseType("Float -> Float"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value.Reduce().(*runtime.Float).Value
			return &runtime.Float{Value: -x}
		},
	})
	env.addFunc("inv", &internalFunc{
		Type: parseType("Float -> Float"), Arity: 1,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Value.Reduce().(*runtime.Float).Value
			return &runtime.Float{Value: 1 / x}
		},
	})
	env.addFunc("+", &internalFunc{
		Type: parseType("Float -> Float -> Float"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return &runtime.Float{Value: x + y}
		},
	})
	env.addFunc("-", &internalFunc{
		Type: parseType("Float -> Float -> Float"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return &runtime.Float{Value: x - y}
		},
	})
	env.addFunc("*", &internalFunc{
		Type: parseType("Float -> Float -> Float"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return &runtime.Float{Value: x * y}
		},
	})
	env.addFunc("/", &internalFunc{
		Type: parseType("Float -> Float -> Float"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return &runtime.Float{Value: x / y}
		},
	})
	env.addFunc("^", &internalFunc{
		Type: parseType("Float -> Float -> Float"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return &runtime.Float{Value: math.Pow(x, y)}
		},
	})
	env.addFunc("==", &internalFunc{
		Type: parseType("Float -> Float -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return boolToRuntimeValue(x == y)
		},
	})
	env.addFunc("!=", &internalFunc{
		Type: parseType("Float -> Float -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return boolToRuntimeValue(x != y)
		},
	})
	env.addFunc("<", &internalFunc{
		Type: parseType("Float -> Float -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return boolToRuntimeValue(x < y)
		},
	})
	env.addFunc("<=", &internalFunc{
		Type: parseType("Float -> Float -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return boolToRuntimeValue(x <= y)
		},
	})
	env.addFunc(">", &internalFunc{
		Type: parseType("Float -> Float -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return boolToRuntimeValue(x > y)
		},
	})
	env.addFunc(">=", &internalFunc{
		Type: parseType("Float -> Float -> Bool"), Arity: 2,
		GoFunc: func(d *runtime.Data) runtime.Value {
			x := d.Next.Value.Reduce().(*runtime.Float).Value
			y := d.Value.Reduce().(*runtime.Float).Value
			return boolToRuntimeValue(x >= y)
		},
	})

	//TODO: useful math functions for Float (sin, cos, sqrt, ...)
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
		return env.addFunc(d.Name, &exprFunc{value})
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
	constructorArity := len(record.Fields)
	err := env.addFunc(
		name,
		&internalFunc{
			SI:    record.SourceInfo(),
			Type:  constructorType,
			Arity: constructorArity,
			GoFunc: func(d *runtime.Data) runtime.Value {
				fields := make([]runtime.Value, constructorArity)
				for i := len(fields) - 1; i >= 0; i-- {
					fields[i] = d.Value
					d = d.Next
				}
				return &runtime.Record{Fields: fields}
			},
		},
	)
	if err != nil {
		return err
	}

	// add record field getters
	// RecordType -> FieldType
	for i, field := range record.Fields {
		index := i
		err := env.addFunc(field.Name, &internalFunc{
			SI:    field.SI,
			Type:  &types.Func{From: recordType, To: field.Type},
			Arity: 1,
			GoFunc: func(d *runtime.Data) runtime.Value {
				r := d.Value.Reduce().(*runtime.Record)
				return r.Fields[index].Reduce()
			},
		})
		if err != nil {
			return err
		}
	}

	// add record field setters
	// (FieldType -> FieldType) -> RecordType -> RecordType
	for i, field := range record.Fields {
		index := i
		err := env.addFunc(field.Name, &internalFunc{
			SI: field.SI,
			Type: &types.Func{
				From: &types.Func{From: field.Type, To: field.Type},
				To:   &types.Func{From: recordType, To: recordType},
			},
			Arity: 2,
			GoFunc: func(d *runtime.Data) runtime.Value {
				f, r := d.Next.Value.Reduce(), d.Value.Reduce()
				oldFields := r.(*runtime.Record).Fields
				newFields := make([]runtime.Value, len(oldFields))
				copy(newFields, oldFields)
				thunk := f.(*runtime.Thunk)
				if thunk.Code.Kind != runtime.CodeAbst {
					panic("not an abstraction")
				}
				newFields[index] = &runtime.Thunk{
					Code: thunk.Code.A,
					Data: runtime.Cons(oldFields[index], runtime.Drop(thunk.Code.Drop, thunk.Data)),
				}
				return &runtime.Record{Fields: newFields}
			},
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
		alternative := i
		var altType types.Type = unionType
		for i := len(alt.Fields) - 1; i >= 0; i-- {
			altType = &types.Func{
				From: alt.Fields[i],
				To:   altType,
			}
		}
		altArity := len(alt.Fields)
		err := env.addFunc(
			alt.Name,
			&internalFunc{
				SI:    alt.SI,
				Type:  altType,
				Arity: altArity,
				GoFunc: func(d *runtime.Data) runtime.Value {
					fields := make([]runtime.Value, altArity)
					for i := len(fields) - 1; i >= 0; i-- {
						fields[i] = d.Value
						d = d.Next
					}
					return &runtime.Union{Alternative: alternative, Fields: fields}
				},
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
