package typecheck

import "github.com/faiface/funky/types"

func Unify(t, u types.Type) (Subst, bool) {
	if v2, ok := u.(*types.Var); ok {
		if v1, ok := t.(*types.Var); !ok || lesserName(v1.Name, v2.Name) {
			return Unify(u, t)
		}
	}

	switch t := t.(type) {
	case *types.Var:
		if _, ok := u.(*types.Var); !ok && ContainsVar(t.Name, u) {
			// occurence check fail
			// variable t is contained in the type u
			// final type would have to be infinitely recursive
			return nil, false
		}
		return Subst{t.Name: u}, true

	case *types.Appl:
		applU, ok := u.(*types.Appl)
		if !ok || t.Cons != applU.Cons || len(t.Args) != len(applU.Args) {
			// t and u are applications of different constructors
			return nil, false
		}
		s := Subst(nil)
		for i := range t.Args {
			// unify application arguments one by one
			// while composing the final substitution
			s1, ok := Unify(s.ApplyToType(t.Args[i]), s.ApplyToType(applU.Args[i]))
			if !ok {
				return nil, false
			}
			s = s.Compose(s1)
		}
		return s, true

	case *types.Func:
		funcU, ok := u.(*types.Func)
		if !ok {
			return nil, false
		}
		s1, ok := Unify(t.From, funcU.From)
		if !ok {
			return nil, false
		}
		s2, ok := Unify(s1.ApplyToType(t.To), s1.ApplyToType(funcU.To))
		if !ok {
			return nil, false
		}
		return s1.Compose(s2), true
	}

	panic("unreachable")
}

func lesserName(s, t string) bool {
	if len(s) < len(t) {
		return true
	}
	if len(s) > len(t) {
		return false
	}
	return s < t
}
