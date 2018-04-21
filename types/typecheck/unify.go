package typecheck

import (
	"github.com/faiface/funky/types"
)

func CheckIfUnify(names map[string]types.Name, t, u types.Type) bool {
	varIndex := 0
	t = instType(&varIndex, t)
	u = instType(&varIndex, u)
	_, ok := Unify(names, t, u)
	return ok
}

func Unify(names map[string]types.Name, t, u types.Type) (Subst, bool) {
	if v2, ok := u.(*types.Var); ok {
		if v1, ok := t.(*types.Var); !ok || lesserName(v1.Name, v2.Name) {
			return Unify(names, u, t)
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
		if !ok || t.Name != applU.Name || len(t.Args) != len(applU.Args) {
			if alias, ok := names[t.Name].(*types.Alias); ok {
				return Unify(names, revealAlias(alias, t.Args), u)
			}
			if ok {
				if alias, ok := names[applU.Name].(*types.Alias); ok {
					return Unify(names, t, revealAlias(alias, applU.Args))
				}
			}
			return nil, false
		}
		s := Subst(nil)
		for i := range t.Args {
			// unify application arguments one by one
			// while composing the final substitution
			s1, ok := Unify(names, s.ApplyToType(t.Args[i]), s.ApplyToType(applU.Args[i]))
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
		s1, ok := Unify(names, t.From, funcU.From)
		if !ok {
			return nil, false
		}
		s2, ok := Unify(names, s1.ApplyToType(t.To), s1.ApplyToType(funcU.To))
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

func revealAlias(alias *types.Alias, args []types.Type) types.Type {
	s := make(Subst)
	for i := range alias.Args {
		s[alias.Args[i]] = args[i]
	}
	return s.ApplyToType(alias.Type)
}
