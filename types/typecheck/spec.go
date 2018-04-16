package typecheck

import "github.com/faiface/funky/types"

func IsSpec(t, u types.Type) bool {
	return isSpec(make(map[string]types.Type), t, u)
}

func isSpec(bind map[string]types.Type, t, u types.Type) bool {
	switch t := t.(type) {
	case *types.Var:
		if bind[t.Name] == nil {
			bind[t.Name] = u
		}
		return bind[t.Name].Equal(u)
	case *types.Appl:
		ua, ok := u.(*types.Appl)
		if !ok || t.Name != ua.Name || len(t.Args) != len(ua.Args) {
			return false
		}
		for i := range t.Args {
			if !isSpec(bind, t.Args[i], ua.Args[i]) {
				return false
			}
		}
		return true
	case *types.Func:
		uf, ok := u.(*types.Func)
		if !ok {
			return false
		}
		return isSpec(bind, t.From, uf.From) && isSpec(bind, t.To, uf.To)
	}
	panic("unreachable")
}
