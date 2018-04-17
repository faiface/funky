package typecheck

import "github.com/faiface/funky/types"

type Vars map[string]types.Type

func (vars Vars) Assume(v string, t types.Type) Vars {
	newVars := make(Vars)
	for v, t := range vars {
		newVars[v] = t
	}
	newVars[v] = t
	return newVars
}

type Funcs map[string][]types.Type

func (fns Funcs) Add(name string, t types.Type) {
	fns[name] = append(fns[name], t)
}
