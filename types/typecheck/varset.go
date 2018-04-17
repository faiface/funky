package typecheck

import (
	"sort"

	"github.com/faiface/funky/types"
)

type VarSet map[string]bool

func (vs VarSet) Put(s string)    { vs[s] = true }
func (vs VarSet) Delete(s string) { delete(vs, s) }

func (vs VarSet) PutAll(ws VarSet) {
	for v := range ws {
		vs.Put(v)
	}
}
func (vs VarSet) DeleteAll(ws VarSet) {
	for v := range ws {
		vs.Delete(v)
	}
}

func (vs VarSet) Copy() VarSet {
	ws := make(VarSet)
	ws.PutAll(vs)
	return ws
}

func (vs VarSet) InOrder() []string {
	var vars []string
	for v := range vs {
		vars = append(vars, v)
	}
	sort.Slice(vars, func(i, j int) bool {
		if len(vars[i]) < len(vars[j]) {
			return true
		}
		if len(vars[i]) > len(vars[j]) {
			return false
		}
		return vars[i] < vars[j]
	})
	return vars
}

func FreeVars(t types.Type) VarSet {
	vs := make(VarSet)
	t.Map(func(t types.Type) types.Type {
		if v, ok := t.(*types.Var); ok {
			vs.Put(v.Name)
		}
		return t
	})
	return vs
}

func ContainsVar(name string, t types.Type) bool {
	contains := false
	t.Map(func(t types.Type) types.Type {
		if v, ok := t.(*types.Var); ok && v.Name == name {
			contains = true
		}
		return t
	})
	return contains
}
