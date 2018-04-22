package compile

import (
	"fmt"

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
	env.names = map[string]types.Name{
		"Char":  &types.Builtin{NumArgs: 0},
		"Int":   &types.Builtin{NumArgs: 0},
		"Float": &types.Builtin{NumArgs: 0},
	}
	env.funcs = make(map[string][]impl)
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
