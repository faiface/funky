package compile

import (
	"github.com/faiface/funky/types"
	"github.com/faiface/funky/types/typecheck"
)

func (env *Env) TypeInfer() []error {
	var errs []error

	global := make(map[string][]types.Type)
	for name, impls := range env.funcs {
		for _, imp := range impls {
			global[name] = append(global[name], imp.TypeInfo())
		}
	}

	for _, impls := range env.funcs {
		for _, imp := range impls {
			impExpr, ok := imp.(*implExpr)
			if !ok {
				continue
			}
			results, err := typecheck.Infer(env.names, global, impExpr.Expr)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			// there's exactly one result
			impExpr.Expr = results[0].Expr
		}
	}

	return errs
}
