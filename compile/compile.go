package compile

import (
	"fmt"

	"github.com/faiface/crux"
	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/runtime"
	"github.com/faiface/funky/types/typecheck"
)

func (env *Env) Compile(main string) (*runtime.Value, error) {
	env.lazyInit()

	globals := make(map[string][]crux.Expr)

	for name, impls := range env.funcs {
		for i := range impls {
			switch impl := impls[i].(type) {
			case *internal:
				globals[name] = append(globals[name], impl.Expr)
			case *function:
				globals[name] = append(globals[name], compress(lift(nil, compress(env.translate(nil, impl.Expr)))))
			}
		}
	}

	if len(globals[main]) == 0 {
		return nil, &Error{
			SourceInfo: nil,
			Msg:        fmt.Sprintf("no %s function", main),
		}
	}
	if len(globals[main]) > 1 {
		return nil, &Error{
			SourceInfo: nil,
			Msg:        fmt.Sprintf("multiple %s functions", main),
		}
	}

	indices, values, _ := crux.Compile(globals)

	return &runtime.Value{Globals: values, Value: values[indices[main][0]]}, nil
}

func (env *Env) translate(locals []string, e expr.Expr) crux.Expr {
	switch e := e.(type) {
	case *expr.Char:
		return &crux.Char{Value: e.Value}

	case *expr.Int:
		var i crux.Int
		i.Value.Set(e.Value)
		return &i

	case *expr.Float:
		return &crux.Float{Value: e.Value}

	case *expr.Var:
		for _, local := range locals {
			if local == e.Name {
				return &crux.Var{Name: e.Name, Index: -1}
			}
		}
		for i, impl := range env.funcs[e.Name] {
			if typecheck.CheckIfUnify(env.names, e.TypeInfo(), impl.TypeInfo()) {
				return &crux.Var{
					Name:  e.Name,
					Index: int32(i),
				}
			}
		}
		panic("unknown variable")

	case *expr.Abst:
		return &crux.Abst{
			Bound: []string{e.Bound.Name},
			Body:  env.translate(append(locals, e.Bound.Name), e.Body),
		}

	case *expr.Appl:
		return &crux.Appl{
			Rator: env.translate(locals, e.Left),
			Rands: []crux.Expr{env.translate(locals, e.Right)},
		}

	case *expr.Strict:
		return &crux.Strict{Expr: env.translate(locals, e.Expr)}

	case *expr.Switch:
		cases := make([]crux.Expr, len(e.Cases))
		for i := range cases {
			cases[i] = env.translate(locals, e.Cases[i].Body)
		}
		return &crux.Switch{
			Expr:  env.translate(locals, e.Expr),
			Cases: cases,
		}

	default:
		panic("unreachable")
	}
}

func compress(e crux.Expr) crux.Expr {
	switch e := e.(type) {
	case *crux.Char, *crux.Int, *crux.Float, *crux.Operator, *crux.Make, *crux.Field, *crux.Var:
		return e

	case *crux.Abst:
		compressedBody := compress(e.Body)
		if abst, ok := compressedBody.(*crux.Abst); ok {
			return &crux.Abst{Bound: append(e.Bound, abst.Bound...), Body: abst.Body}
		}
		return &crux.Abst{Bound: e.Bound, Body: compressedBody}

	case *crux.Appl:
		compressedRator := compress(e.Rator)
		compressedRands := make([]crux.Expr, len(e.Rands))
		for i := range compressedRands {
			compressedRands[i] = compress(e.Rands[i])
		}
		if appl, ok := compressedRator.(*crux.Appl); ok {
			return &crux.Appl{Rator: appl.Rator, Rands: append(appl.Rands, compressedRands...)}
		}
		return &crux.Appl{Rator: compressedRator, Rands: compressedRands}

	case *crux.Strict:
		return &crux.Strict{Expr: compress(e.Expr)}

	case *crux.Switch:
		compressedExpr := compress(e.Expr)
		compressedCases := make([]crux.Expr, len(e.Cases))
		for i := range compressedCases {
			compressedCases[i] = compress(e.Cases[i])
		}
		return &crux.Switch{Expr: compressedExpr, Cases: compressedCases}

	default:
		panic("unreachable")
	}
}

func lift(locals []string, e crux.Expr) crux.Expr {
	switch e := e.(type) {
	case *crux.Char, *crux.Int, *crux.Float, *crux.Operator, *crux.Make, *crux.Field, *crux.Var:
		return e

	case *crux.Abst:
		var (
			toBind  []string
			toApply []crux.Expr
		)
		for _, local := range locals {
			if isFree(local, e) {
				toBind = append(toBind, local)
				toApply = append(toApply, &crux.Var{Name: local, Index: -1})
			}
		}
		if len(toBind) == 0 {
			return &crux.Abst{Bound: e.Bound, Body: lift(e.Bound, e.Body)}
		}
		newLocals := append(toBind, e.Bound...)
		return &crux.Appl{
			Rator: &crux.Abst{Bound: newLocals, Body: lift(newLocals, e.Body)},
			Rands: toApply,
		}

	case *crux.Appl:
		liftedRator := lift(locals, e.Rator)
		liftedRands := make([]crux.Expr, len(e.Rands))
		for i := range liftedRands {
			liftedRands[i] = lift(locals, e.Rands[i])
		}
		return &crux.Appl{Rator: liftedRator, Rands: liftedRands}

	case *crux.Strict:
		return &crux.Strict{Expr: lift(locals, e.Expr)}

	case *crux.Switch:
		liftedExpr := lift(locals, e.Expr)
		liftedCases := make([]crux.Expr, len(e.Cases))
		for i := range liftedCases {
			liftedCases[i] = lift(locals, e.Cases[i])
		}
		return &crux.Switch{Expr: liftedExpr, Cases: liftedCases}

	default:
		panic("unreachable")
	}
}

func isFree(local string, e crux.Expr) bool {
	switch e := e.(type) {
	case *crux.Char, *crux.Int, *crux.Float, *crux.Operator, *crux.Make, *crux.Field:
		return false
	case *crux.Var:
		return e.Index < 0 && local == e.Name
	case *crux.Abst:
		for _, bound := range e.Bound {
			if local == bound {
				return false
			}
		}
		return isFree(local, e.Body)
	case *crux.Appl:
		for _, rand := range e.Rands {
			if isFree(local, rand) {
				return true
			}
		}
		return isFree(local, e.Rator)
	case *crux.Strict:
		return isFree(local, e.Expr)
	case *crux.Switch:
		for _, cas := range e.Cases {
			if isFree(local, cas) {
				return true
			}
		}
		return isFree(local, e.Expr)
	default:
		panic("unreachable")
	}
}
