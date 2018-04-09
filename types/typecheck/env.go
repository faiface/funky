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

type Defs map[string][]types.Type

func (defs Defs) Define(name string, t types.Type) {
	defs[name] = append(defs[name], t)
}
