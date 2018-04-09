package typecheck

import (
	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/types"
)

type Subst map[string]types.Type

func (s Subst) Compose(s1 Subst) Subst {
	s2 := make(Subst)
	for v, t := range s { // copy s + transitivity
		s2[v] = s1.ApplyToType(t)
	}
	for v, t := range s1 { // copy s1
		s2[v] = t
	}
	return s2
}

func (s Subst) ApplyToType(t types.Type) types.Type {
	if t == nil {
		return nil
	}
	return t.Map(func(t types.Type) types.Type {
		if v, ok := t.(*types.Var); ok && s[v.Name] != nil {
			return s[v.Name]
		}
		return t
	})
}

func (s Subst) ApplyToExpr(e expr.Expr) expr.Expr {
	if e == nil {
		return nil
	}
	return e.Map(func(e expr.Expr) expr.Expr {
		return e.WithTypeInfo(s.ApplyToType(e.TypeInfo()))
	})
}

func (s Subst) ApplyToVars(env Vars) Vars {
	newEnv := make(Vars)
	for v, t := range env {
		newEnv[v] = s.ApplyToType(t)
	}
	return newEnv
}
