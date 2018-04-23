package compile

import (
	"fmt"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/runtime"
	"github.com/faiface/funky/types"
	"github.com/faiface/funky/types/typecheck"
)

func (env *Env) Compile(main string) (*runtime.Value, error) {
	if len(env.funcs[main]) == 0 {
		return nil, &Error{
			SourceInfo: nil,
			Msg:        fmt.Sprintf("no %s function", main),
		}
	}
	if len(env.funcs[main]) > 1 {
		return nil, &Error{
			SourceInfo: nil,
			Msg:        fmt.Sprintf("multiple %s functions", main),
		}
	}

	offsets := make(map[string]int)
	var exprs []runtime.Expr

	// allocate space for compiled expressions
	for name, impls := range env.funcs {
		offsets[name] = len(exprs)
		for range impls {
			exprs = append(exprs, nil)
		}
	}

	// gather function type info
	global := make(map[string][]types.Type)
	for name, impls := range env.funcs {
		for _, imp := range impls {
			global[name] = append(global[name], imp.TypeInfo())
		}
	}

	// compile individual functions
	for name, impls := range env.funcs {
		for i, imp := range impls {
			switch imp := imp.(type) {
			case *implInternal:
				exprs[offsets[name]+i] = imp.Expr
			case *implExpr:
				exprs[offsets[name]+i] = compile(env.names, offsets, exprs, global, nil, imp.Expr)
			}
		}
	}

	return &runtime.Value{Expr: exprs[offsets[main]].WithCtx(nil)}, nil
}

func compile(
	names map[string]types.Name,
	offsets map[string]int,
	exprs []runtime.Expr,
	global map[string][]types.Type,
	args []string,
	e expr.Expr,
) runtime.Expr {
	switch e := e.(type) {
	case *expr.Var:
		for i, arg := range args {
			if arg == e.Name {
				return runtime.Var{Index: i}
			}
		}
		for i, typ := range global[e.Name] {
			if typecheck.CheckIfUnify(names, e.TypeInfo(), typ) {
				return runtime.Ref{
					Expr: &exprs[offsets[e.Name]+i],
				}
			}
		}
		panic("unreachable")

	case *expr.Appl:
		ldrop := 0
		for _, name := range args {
			if e.Left.HasFree(name) {
				break
			}
			ldrop++
		}
		rdrop := 0
		for _, name := range args {
			if e.Right.HasFree(name) {
				break
			}
			rdrop++
		}
		largs, rargs := args[ldrop:], args[rdrop:]

		left := compile(names, offsets, exprs, global, largs, e.Left)
		right := compile(names, offsets, exprs, global, rargs, e.Right)

		return &runtime.Appl{
			LDrop: ldrop,
			RDrop: rdrop,
			Left:  left,
			Right: right,
		}

	case *expr.Abst:
		return &runtime.Abst{
			Body: compile(names, offsets, exprs, global, append([]string{e.Bound.Name}, args...), e.Body),
		}

	case *expr.Switch:
		cases := make([]runtime.Expr, len(e.Cases))
		for i := range cases {
			cases[i] = compile(names, offsets, exprs, global, args, e.Cases[i].Body)
		}
		return runtime.Switch{
			Expr:  compile(names, offsets, exprs, global, args, e.Expr),
			Cases: cases,
		}

	case *expr.Char:
		return runtime.Char(e.Value)

	case *expr.Int:
		return (*runtime.Int)(e.Value)

	case *expr.Float:
		return runtime.Float(e.Value)
	}

	panic("unreachable")
}
