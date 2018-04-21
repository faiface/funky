package typecheck

import "github.com/faiface/funky/types"

func IsSpec(names map[string]types.Name, t, u types.Type) bool {
	return isSpec(names, make(map[string]types.Type), t, u)
}

func isSpec(names map[string]types.Name, bind map[string]types.Type, t, u types.Type) bool {
	switch t := t.(type) {
	case *types.Var:
		if bind[t.Name] == nil {
			bind[t.Name] = u
		}
		return bind[t.Name].Equal(u)
	case *types.Appl:
		ua, ok := u.(*types.Appl)
		if !ok || t.Name != ua.Name || len(t.Args) != len(ua.Args) {
			if alias, ok := names[t.Name].(*types.Alias); ok {
				return isSpec(names, bind, revealAlias(alias, t.Args), u)
			}
			if ok {
				if alias, ok := names[ua.Name].(*types.Alias); ok {
					return isSpec(names, bind, t, revealAlias(alias, ua.Args))
				}
			}
			return false
		}
		for i := range t.Args {
			if !isSpec(names, bind, t.Args[i], ua.Args[i]) {
				return false
			}
		}
		return true
	case *types.Func:
		uf, ok := u.(*types.Func)
		if !ok {
			return false
		}
		return isSpec(names, bind, t.From, uf.From) && isSpec(names, bind, t.To, uf.To)
	}
	panic("unreachable")
}
