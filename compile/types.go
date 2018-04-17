package compile

import (
	"fmt"

	"github.com/faiface/funky/types"
	"github.com/faiface/funky/types/typecheck"
)

func (env *Env) Validate() []error {
	var errs []error

	for _, definition := range env.names {
		switch definition := definition.(type) {
		case *types.Builtin:
			continue
		case *types.Record:
			err := env.validateRecord(definition)
			if err != nil {
				errs = append(errs, err)
			}
		default:
			panic("unreachable")
		}
	}

	for _, impls := range env.funcs {
		for _, imp := range impls {
			freeVars := typecheck.FreeVars(imp.TypeInfo()).InOrder()
			err := env.validateType(freeVars, imp.TypeInfo())
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func (env *Env) validateType(boundVars []string, typ types.Type) error {
	switch typ := typ.(type) {
	case *types.Var:
		for _, bound := range boundVars {
			if typ.Name == bound {
				return nil
			}
		}
		return &Error{typ.SourceInfo(), fmt.Sprintf("type variable not bound: %s", typ.Name)}

	case *types.Appl:
		if env.names[typ.Name] == nil {
			return &Error{typ.SourceInfo(), fmt.Sprintf("type name does not exist: %s", typ.Name)}
		}
		numArgs := len(typ.Args)
		arity := env.names[typ.Name].Arity()
		if numArgs != arity {
			return &Error{
				typ.SourceInfo(),
				fmt.Sprintf("type %s requires %d arguments, %d given", typ.Name, arity, numArgs),
			}
		}
		for _, arg := range typ.Args {
			err := env.validateType(boundVars, arg)
			if err != nil {
				return err
			}
		}
		return nil

	case *types.Func:
		err := env.validateType(boundVars, typ.From)
		if err != nil {
			return err
		}
		err = env.validateType(boundVars, typ.To)
		if err != nil {
			return err
		}
		return nil
	}

	panic("unreachable")
}

func (env *Env) validateRecord(record *types.Record) error {
	for _, field := range record.Fields {
		err := env.validateType(record.Args, field.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func (env *Env) TypeInfer() []error {
	var errs []error

	global := make(typecheck.Funcs)
	for name, impls := range env.funcs {
		for _, imp := range impls {
			global.Add(name, imp.TypeInfo())
		}
	}

	for _, impls := range env.funcs {
		for _, imp := range impls {
			impExpr, ok := imp.(*implExpr)
			if !ok {
				continue
			}
			results, err := typecheck.Infer(global, impExpr.expr)
			if err != nil {
				errs = append(errs, err)
			}
			if err == nil {
				// there's exactly one result
				impExpr.expr = results[0].Expr
			}
		}
	}

	return errs
}
