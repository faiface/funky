package typecheck

import (
	"fmt"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/parse/parseinfo"
	"github.com/faiface/funky/types"
)

type NotBoundError struct {
	SourceInfo *parseinfo.Source
	Name       string
}

func (err *NotBoundError) Error() string {
	return fmt.Sprintf("%v: variable not bound: %s", err.SourceInfo, err.Name)
}

type CannotApplyError struct {
	Cases []struct {
		Left  expr.Expr
		Right expr.Expr
		Err   error
	}
}

func (err *CannotApplyError) Error() string {
	return "cannot apply"
}

type NoMatchErr struct {
	Results []InferResult
}

func (err *NoMatchErr) Error() string {
	return "no match"
}

type InferResult struct {
	Type  types.Type
	Subst Subst
	Expr  expr.Expr
}

func Infer(global Defs, e expr.Expr) ([]InferResult, error) {
	varIndex := 0
	e = instExpr(&varIndex, e)
	results, err := infer(&varIndex, global, make(Vars), e)
	if err != nil {
		return nil, err
	}
	for i := range results {
		results[i].Expr = results[i].Subst.ApplyToExpr(results[i].Expr)
	}
	return results, nil
}

func infer(varIndex *int, global Defs, local Vars, e expr.Expr) (results []InferResult, err error) {
	defer func() {
		if err != nil || e.TypeInfo() == nil {
			return
		}
		// filter infer results by the type info
		var filtered []InferResult
		for _, r := range results {
			if IsSpec(r.Type, e.TypeInfo()) {
				s, _ := Unify(r.Type, e.TypeInfo())
				r.Type = s.ApplyToType(r.Type)
				r.Subst = r.Subst.Compose(s)
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			err = &NoMatchErr{results}
			results = nil
			return
		}
		results = filtered
		err = nil
	}()

	switch e := e.(type) {
	case *expr.Var:
		if t, ok := local[e.Name]; ok {
			return []InferResult{{
				Type:  t,
				Subst: nil,
				Expr:  e.WithTypeInfo(t),
			}}, nil
		}
		if ts, ok := global[e.Name]; ok {
			results = nil
			for _, t := range ts {
				t = instType(varIndex, t)
				results = append(results, InferResult{
					Type:  t,
					Subst: nil,
					Expr:  e.WithTypeInfo(t),
				})
			}
			return results, nil
		}
		return nil, &NotBoundError{e.SourceInfo(), e.Name}

	case *expr.Appl:
		results1, err := infer(varIndex, global, local, e.Left)
		if err != nil {
			return nil, err
		}
		{ // if the right side is wrong in itself, return a simple error from there
			_, err := infer(varIndex, global, local, e.Right)
			if err != nil {
				return nil, err
			}
		}
		results = nil
		cannotApplyErr := new(CannotApplyError)
		for _, r1 := range results1 {
			results2, err := infer(
				varIndex,
				global,
				r1.Subst.ApplyToVars(local),
				e.Right,
			)
			if err != nil {
				cannotApplyErr.Cases = append(cannotApplyErr.Cases, struct {
					Left  expr.Expr
					Right expr.Expr
					Err   error
				}{r1.Expr, nil, err})
			}
			resultType := newVar(varIndex)
			for _, r2 := range results2 {
				s, ok := Unify(
					r2.Subst.ApplyToType(r1.Type),
					&types.Func{
						From: r2.Type,
						To:   resultType,
					},
				)
				if !ok {
					cannotApplyErr.Cases = append(cannotApplyErr.Cases, struct {
						Left  expr.Expr
						Right expr.Expr
						Err   error
					}{r1.Expr, r2.Expr, nil})
					continue
				}
				t := s.ApplyToType(resultType)
				results = append(results, InferResult{
					Type:  t,
					Subst: r1.Subst.Compose(r2.Subst).Compose(s),
					Expr: &expr.Appl{
						TI:    t,
						Left:  r1.Expr,
						Right: r2.Expr,
					},
				})
			}
		}
		if len(results) == 0 {
			return nil, cannotApplyErr
		}
		return results, nil

	case *expr.Abst:
		var bindType types.Type
		if f, ok := e.TypeInfo().(*types.Func); ok {
			bindType = f.From
		} else {
			bindType = newVar(varIndex)
		}
		newLocal := local.Assume(e.Bound.Name, bindType)
		bodyResults, err := infer(varIndex, global, newLocal, e.Body)
		if err != nil {
			return nil, err
		}
		results = nil
		for _, r := range bodyResults {
			inferredBindType := r.Subst.ApplyToType(bindType)
			t := &types.Func{
				From: inferredBindType,
				To:   r.Type,
			}
			results = append(results, InferResult{
				Type:  t,
				Subst: r.Subst,
				Expr: &expr.Abst{
					TI:    t,
					Bound: e.Bound.WithTypeInfo(inferredBindType).(*expr.Var),
					Body:  r.Expr,
				},
			})
		}
		return results, nil
	}

	panic("unreachable")
}

func newVar(varIndex *int) *types.Var {
	name := ""
	i := *varIndex + 1
	for i > 0 {
		name = string('a'+(i-1)%26) + name
		i = (i - 1) / 26
	}
	v := &types.Var{Name: name}
	*varIndex++
	return v
}

func instTypeHelper(varIndex *int, renames map[string]string, t types.Type) types.Type {
	return t.Map(func(t types.Type) types.Type {
		if v, ok := t.(*types.Var); ok {
			renamed, ok := renames[v.Name]
			if !ok {
				renamed = newVar(varIndex).Name
				renames[v.Name] = renamed
				*varIndex++
			}
			return &types.Var{
				SI:   v.SI,
				Name: renamed,
			}
		}
		return t
	})
}

func instType(varIndex *int, t types.Type) types.Type {
	renames := make(map[string]string)
	return instTypeHelper(varIndex, renames, t)
}

func instExpr(varIndex *int, e expr.Expr) expr.Expr {
	renames := make(map[string]string)
	return e.Map(func(e expr.Expr) expr.Expr {
		t := e.TypeInfo()
		if t != nil {
			t = instTypeHelper(varIndex, renames, t)
		}
		return e.WithTypeInfo(t)
	})
}
