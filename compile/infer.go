package compile

import (
	"github.com/faiface/funky/types"
	"github.com/faiface/funky/types/typecheck"
)

func (env *Env) TypeInfer() []error {
	env.lazyInit()

	var errs []error

	global := make(map[string][]types.Type)
	for name, impls := range env.funcs {
		for _, imp := range impls {
			global[name] = append(global[name], imp.TypeInfo())
		}
	}

	for _, impls := range env.funcs {
		for _, imp := range impls {
			function, ok := imp.(*function)
			if !ok {
				continue
			}
			results, err := typecheck.Infer(env.names, global, function.Expr)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			// there's exactly one result
			function.Expr = results[0].Expr
		}
	}

	return errs
}
