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
	Left, Right expr.Expr
}

func (err *CannotApplyError) Error() string {
	return fmt.Sprintf(
		"%v: cannot apply (%v) to (%v)",
		err.Left.SourceInfo(),
		err.Left.TypeInfo(),
		err.Right.TypeInfo(),
	)
}

type InferResult struct {
	Type  types.Type
	Subst Subst
	Expr  expr.Expr
	Err   error
}

func Infer(global Defs, e expr.Expr) ([]InferResult, error) {
	varIndex := 0
	e = instExpr(&varIndex, e)
	return infer(&varIndex, global, make(Vars), e)
}

func infer(varIndex *int, global Defs, local Vars, e expr.Expr) (results []InferResult, err error) {
	typeInfo := e.TypeInfo()
	if typeInfo != nil {
		defer func() {
			//TODO: check if matches results and filter accordingly
		}()
	}

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
		for _, r1 := range results1 {
			if r1.Err != nil {
				results = append(results, InferResult{Err: r1.Err})
				continue
			}
			results2, err := infer(
				varIndex,
				global,
				r1.Subst.ApplyToVars(local),
				e.Right,
			)
			if err != nil {
				panic("unreachable")
			}
			resultType := newVar(varIndex)
			for _, r2 := range results2 {
				if r2.Err != nil {
					results = append(results, InferResult{Err: r2.Err})
					continue
				}
				s, ok := Unify(
					r2.Subst.ApplyToType(r1.Type),
					&types.Func{
						From: r2.Type,
						To:   resultType,
					},
				)
				if !ok {
					results = append(results, InferResult{
						Err: &CannotApplyError{r1.Expr, r2.Expr},
					})
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
		return results, nil

	case *expr.Abst:
		bindType := newVar(varIndex)
		newLocal := local.Assume(e.Bound.Name, bindType)
		bodyResults, err := infer(varIndex, global, newLocal, e.Body)
		if err != nil {
			return nil, err
		}
		results = nil
		for _, r := range bodyResults {
			if r.Err != nil {
				results = append(results, InferResult{Err: r.Err})
				continue
			}
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
