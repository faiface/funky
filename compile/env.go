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
	funcs  map[string][]impl
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

	env.funcs = make(map[string][]impl)

	// built-in functions

	// common
	env.addFunc("eval", &implInternal{
		Type: parseType("a -> b -> b"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			args[0].Reduce()
			return args[1]
		}),
	})
	env.addFunc("dump", &implInternal{
		Type: parseType("String -> a -> a"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			var builder strings.Builder
			union := args[0].Reduce().(runtime.Union)
			for union.Alternative != 0 {
				builder.WriteRune(rune(union.Fields[0].Reduce().(runtime.Char)))
				union = union.Fields[1].Reduce().(runtime.Union)
			}
			fmt.Fprint(os.Stderr, builder.String())
			return args[1]
		}),
	})
	env.addFunc("error", &implInternal{
		Type: parseType("String -> a"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			var builder strings.Builder
			union := args[0].Reduce().(runtime.Union)
			for union.Alternative != 0 {
				builder.WriteRune(rune(union.Fields[0].Reduce().(runtime.Char)))
				union = union.Fields[1].Reduce().(runtime.Union)
			}
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", builder.String())
			os.Exit(1)
			return nil
		}),
	})

	// conversions
	env.addFunc("int", &implInternal{
		Type: parseType("Char -> Int"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			return (*runtime.Int)(big.NewInt(int64(args[0].Reduce().(runtime.Char))))
		}),
	})
	env.addFunc("char", &implInternal{
		Type: parseType("Int -> Char"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			return runtime.Char((*big.Int)(args[0].Reduce().(*runtime.Int)).Int64())
		}),
	})
	env.addFunc("int", &implInternal{
		Type: parseType("Float -> Int"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			x := math.Floor(float64(args[0].Reduce().(runtime.Float)))
			z, _ := big.NewFloat(x).Int(nil)
			return (*runtime.Int)(z)
		}),
	})
	env.addFunc("float", &implInternal{
		Type: parseType("Int -> Float"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			z, _ := big.NewFloat(0).SetInt(x).Float64()
			return runtime.Float(z)
		}),
	})

	// Int
	env.addFunc("neg", &implInternal{
		Type: parseType("Int -> Int"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			z := big.NewInt(0)
			z.Neg(x)
			return (*runtime.Int)(z)
		}),
	})
	env.addFunc("+", &implInternal{
		Type: parseType("Int -> Int -> Int"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			z := big.NewInt(0)
			z.Add(x, y)
			return (*runtime.Int)(z)
		}),
	})
	env.addFunc("-", &implInternal{
		Type: parseType("Int -> Int -> Int"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			z := big.NewInt(0)
			z.Sub(x, y)
			return (*runtime.Int)(z)
		}),
	})
	env.addFunc("*", &implInternal{
		Type: parseType("Int -> Int -> Int"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			z := big.NewInt(0)
			z.Mul(x, y)
			return (*runtime.Int)(z)
		}),
	})
	env.addFunc("/", &implInternal{
		Type: parseType("Int -> Int -> Int"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			z := big.NewInt(0)
			z.Div(x, y)
			return (*runtime.Int)(z)
		}),
	})
	env.addFunc("%", &implInternal{
		Type: parseType("Int -> Int -> Int"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			z := big.NewInt(0)
			z.Mod(x, y)
			return (*runtime.Int)(z)
		}),
	})
	env.addFunc("^", &implInternal{
		Type: parseType("Int -> Int -> Int"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			z := big.NewInt(0)
			z.Exp(x, y, nil)
			return (*runtime.Int)(z)
		}),
	})
	env.addFunc("==", &implInternal{
		Type: parseType("Int -> Int -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			return runtime.MkBoolExpr(x.Cmp(y) == 0)
		}),
	})
	env.addFunc("!=", &implInternal{
		Type: parseType("Int -> Int -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			return runtime.MkBoolExpr(x.Cmp(y) != 0)
		}),
	})
	env.addFunc("<", &implInternal{
		Type: parseType("Int -> Int -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			return runtime.MkBoolExpr(x.Cmp(y) < 0)
		}),
	})
	env.addFunc("<=", &implInternal{
		Type: parseType("Int -> Int -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			return runtime.MkBoolExpr(x.Cmp(y) <= 0)
		}),
	})
	env.addFunc(">", &implInternal{
		Type: parseType("Int -> Int -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			return runtime.MkBoolExpr(x.Cmp(y) > 0)
		}),
	})
	env.addFunc(">=", &implInternal{
		Type: parseType("Int -> Int -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			y := (*big.Int)(args[1].Reduce().(*runtime.Int))
			return runtime.MkBoolExpr(x.Cmp(y) >= 0)
		}),
	})
	env.addFunc("string", &implInternal{
		Type: parseType("Int -> String"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			x := (*big.Int)(args[0].Reduce().(*runtime.Int))
			return runtime.MkStringExpr(x.Text(10))
		}),
	})

	// Float
	env.addFunc("neg", &implInternal{
		Type: parseType("Float -> Float"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			return runtime.Float(-x)
		}),
	})
	env.addFunc("inv", &implInternal{
		Type: parseType("Float -> Float"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			return runtime.Float(1 / x)
		}),
	})
	env.addFunc("+", &implInternal{
		Type: parseType("Float -> Float -> Float"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.Float(x + y)
		}),
	})
	env.addFunc("-", &implInternal{
		Type: parseType("Float -> Float -> Float"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.Float(x - y)
		}),
	})
	env.addFunc("*", &implInternal{
		Type: parseType("Float -> Float -> Float"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.Float(x * y)
		}),
	})
	env.addFunc("/", &implInternal{
		Type: parseType("Float -> Float -> Float"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.Float(x / y)
		}),
	})
	env.addFunc("^", &implInternal{
		Type: parseType("Float -> Float -> Float"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.Float(math.Pow(x, y))
		}),
	})
	env.addFunc("==", &implInternal{
		Type: parseType("Float -> Float -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.MkBoolExpr(x == y)
		}),
	})
	env.addFunc("!=", &implInternal{
		Type: parseType("Float -> Float -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.MkBoolExpr(x != y)
		}),
	})
	env.addFunc("<", &implInternal{
		Type: parseType("Float -> Float -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.MkBoolExpr(x < y)
		}),
	})
	env.addFunc("<=", &implInternal{
		Type: parseType("Float -> Float -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.MkBoolExpr(x <= y)
		}),
	})
	env.addFunc(">", &implInternal{
		Type: parseType("Float -> Float -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.MkBoolExpr(x > y)
		}),
	})
	env.addFunc(">=", &implInternal{
		Type: parseType("Float -> Float -> Bool"),
		Expr: makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			y := float64(args[1].Reduce().(runtime.Float))
			return runtime.MkBoolExpr(x >= y)
		}),
	})
	env.addFunc("string", &implInternal{
		Type: parseType("Float -> String"),
		Expr: makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
			x := float64(args[0].Reduce().(runtime.Float))
			return runtime.MkStringExpr(strconv.FormatFloat(x, 'f', -1, 64))
		}),
	})

	//TODO: useful math functions for Float (sin, cos, sqrt, ...)
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
		return env.addFunc(d.Name, &implExpr{value})
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
		&implInternal{
			record.SourceInfo(),
			constructorType,
			makeGoFunc(len(record.Fields), func(args ...runtime.Expr) runtime.Expr {
				return runtime.Record{Fields: args}
			}),
		},
	)
	if err != nil {
		return err
	}

	// add record field getters
	for i, field := range record.Fields {
		index := i
		err := env.addFunc(field.Name, &implInternal{
			field.SI,
			&types.Func{
				From: recordType,
				To:   field.Type,
			},
			makeGoFunc(1, func(args ...runtime.Expr) runtime.Expr {
				return args[0].Reduce().(runtime.Record).Fields[index]
			}),
		})
		if err != nil {
			return err
		}
	}

	// add record fiel setters
	for i, field := range record.Fields {
		index := i
		err := env.addFunc(field.Name, &implInternal{
			field.SI,
			&types.Func{
				From: field.Type,
				To:   &types.Func{From: recordType, To: recordType},
			},
			makeGoFunc(2, func(args ...runtime.Expr) runtime.Expr {
				newFields := make([]runtime.Expr, len(record.Fields))
				copy(newFields, args[1].Reduce().(runtime.Record).Fields)
				newFields[index] = args[0]
				return runtime.Record{Fields: newFields}
			}),
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
		var altType types.Type = unionType
		for i := len(alt.Fields) - 1; i >= 0; i-- {
			altType = &types.Func{
				From: alt.Fields[i],
				To:   altType,
			}
		}
		alternative := i
		err := env.addFunc(
			alt.Name,
			&implInternal{
				alt.SI,
				altType,
				makeGoFunc(len(alt.Fields), func(args ...runtime.Expr) runtime.Expr {
					return runtime.Union{Alternative: alternative, Fields: args}
				}),
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

func (env *Env) addFunc(name string, imp impl) error {
	env.funcs[name] = append(env.funcs[name], imp)
	return nil
}

type argList struct {
	Value runtime.Expr
	Next  *argList
}

func cons(value runtime.Expr, next *argList) *argList {
	return &argList{
		Value: value,
		Next:  next,
	}
}

func makeGoFunc(arity int, fn func(...runtime.Expr) runtime.Expr) runtime.Expr {
	if arity == 0 {
		return fn()
	}
	if arity == 1 {
		return runtime.GoFunc(func(e runtime.Expr) runtime.Expr {
			return fn(e)
		})
	}
	if arity == 2 {
		return runtime.GoFunc(func(e1 runtime.Expr) runtime.Expr {
			return runtime.GoFunc(func(e2 runtime.Expr) runtime.Expr {
				return fn(e1, e2)
			})
		})
	}
	return makeGoFuncHelper(arity, nil, fn)
}

func makeGoFuncHelper(left int, al *argList, fn func(...runtime.Expr) runtime.Expr) runtime.Expr {
	if left == 0 {
		var args []runtime.Expr
		for al != nil {
			args = append(args, al.Value)
			al = al.Next
		}
		for i, j := 0, len(args)-1; i < j; i, j = i+1, j-1 {
			args[i], args[j] = args[j], args[i]
		}
		return fn(args...)
	}
	return runtime.GoFunc(func(e runtime.Expr) runtime.Expr {
		return makeGoFuncHelper(left-1, cons(e, al), fn)
	})
}
