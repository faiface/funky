package typecheck

import "github.com/faiface/funky/types"

type Env map[string]types.Type

func (env Env) Assume(v string, t types.Type) Env {
	newEnv := make(Env)
	for v, t := range env {
		newEnv[v] = t
	}
	newEnv[v] = t
	return newEnv
}
