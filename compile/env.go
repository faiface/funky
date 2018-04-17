package compile

import (
	"fmt"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/parse"
	"github.com/faiface/funky/parse/parseinfo"
	"github.com/faiface/funky/types"
	"github.com/faiface/funky/types/typecheck"
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
	defs   map[string][]impl
}

func (env *Env) lazyInit() {
	if env.inited {
		return
	}
	env.inited = true
	env.names = make(map[string]types.Name)
	env.defs = make(map[string][]impl)
}

func (env *Env) Add(d parse.Definition) error {
	env.lazyInit()

	switch value := d.Value.(type) {
	case *types.Record:
		if env.names[d.Name] != nil {
			return &Error{
				value.SourceInfo(),
				fmt.Sprintf("type name %s already defined: %v", d.Name, env.names[d.Name].SourceInfo()),
			}
		}
		env.names[d.Name] = value

		var args []types.Type
		for _, arg := range value.Args {
			args = append(args, &types.Var{Name: arg})
		}
		recordType := &types.Appl{
			SI:   value.SourceInfo(),
			Name: d.Name,
			Args: args,
		}

		// add record constructor
		var constructorType types.Type = recordType
		for i := len(value.Fields) - 1; i >= 0; i-- {
			constructorType = &types.Func{
				From: value.Fields[i].Type,
				To:   constructorType,
			}
		}
		err := env.addDef(d.Name, &implUndefined{constructorType})
		if err != nil {
			return err
		}

		// add record field getters
		for _, field := range value.Fields {
			err := env.addDef(field.Name, &implUndefined{
				&types.Func{
					From: recordType,
					To:   field.Type,
				},
			})
			if err != nil {
				return err
			}
		}

	case expr.Expr:
		err := env.addDef(d.Name, &implExpr{value})
		if err != nil {
			return err
		}
	}

	return nil
}

func (env *Env) addDef(name string, imp impl) error {
	env.lazyInit()

	for _, alreadyDefined := range env.defs[name] {
		if _, ok := typecheck.Unify(imp.TypeInfo(), alreadyDefined.TypeInfo()); ok {
			return &Error{
				imp.SourceInfo(),
				fmt.Sprintf(
					"definition of %s with colliding signature exists: %v",
					name,
					alreadyDefined.SourceInfo(),
				),
			}
		}
	}

	env.defs[name] = append(env.defs[name], imp)

	return nil
}

func (env *Env) TypeInfer() error {
	global := make(typecheck.Defs)
	for name, impls := range env.defs {
		for _, imp := range impls {
			global.Define(name, imp.TypeInfo())
		}
	}

	for _, impls := range env.defs {
		for _, imp := range impls {
			impExpr, ok := imp.(*implExpr)
			if !ok {
				continue
			}
			results, err := typecheck.Infer(global, impExpr.expr)
			if err != nil {
				return err
			}
			// there's exactly one result
			impExpr.expr = results[0].Expr
		}
	}

	return nil
}
