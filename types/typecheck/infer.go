package typecheck

import (
	"fmt"
	"strings"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/parse/parseinfo"
	"github.com/faiface/funky/types"
)

type (
	NotBoundError struct {
		SourceInfo *parseinfo.Source
		Name       string
	}

	CannotApplyError struct {
		LeftSourceInfo  *parseinfo.Source
		RightSourceInfo *parseinfo.Source
		Cases           []struct {
			Left  types.Type
			Right types.Type
			Err   error
		}
	}

	NoMatchError struct {
		SourceInfo *parseinfo.Source
		TypeInfo   types.Type
		Results    []InferResult
	}

	AmbiguousError struct {
		SourceInfo *parseinfo.Source
		TypeInfo   types.Type
		Results    []InferResult
	}
)

func (err *NotBoundError) Error() string {
	return fmt.Sprintf("%v: variable not bound: %s", err.SourceInfo, err.Name)
}

func (err *CannotApplyError) Error() string {
	s := fmt.Sprintf("%v: cannot apply; in case function has type:", err.LeftSourceInfo)
	for _, cas := range err.Cases {
		s += "\n" + cas.Left.String()
		if cas.Err != nil {
			s += "\n" + indent(cas.Err.Error())
			continue
		}
		s += "\n  " + fmt.Sprintf("%v: argument does not match: %v", err.RightSourceInfo, cas.Right)
	}
	return s
}

func indent(s string) string {
	var b strings.Builder
	if len(s) > 0 {
		b.WriteString("  ")
	}
	for _, r := range s {
		b.WriteRune(r)
		if r == '\n' {
			b.WriteString("  ")
		}
	}
	return b.String()
}

func (err *NoMatchError) Error() string {
	s := fmt.Sprintf("%v: expression type does not match required type: %v\n", err.SourceInfo, err.TypeInfo)
	s += "admissible types are:"
	for _, r := range err.Results {
		s += fmt.Sprintf("\n  %v", r.Type)
	}
	return s
}

func (err *AmbiguousError) Error() string {
	traversals := make([]<-chan expr.Expr, len(err.Results))
	for i := range traversals {
		traversals[i] = traverse(err.Results[i].Subst.ApplyToExpr(err.Results[i].Expr))
	}
	// the idea is to concurrently traverse all inferred expressions and find the first
	// variable that differs in type across the results and report it
	for {
		var exprs []expr.Expr
		for i := range traversals {
			exprs = append(exprs, <-traversals[i])
		}
		for i := 1; i < len(exprs); i++ {
			if !exprs[0].TypeInfo().Equal(exprs[i].TypeInfo()) {
				// we found one source of ambiguity, we report it
				s := fmt.Sprintf("%v: ambiguous, multiple admissible types:", exprs[0].SourceInfo())
			accumulateTypes:
				for j, e := range exprs {
					for k := 0; k < j; k++ {
						if exprs[k].TypeInfo().Equal(e.TypeInfo()) {
							continue accumulateTypes
						}
					}
					s += fmt.Sprintf("\n  %v", e.TypeInfo())
				}
				// drain traversals
				for _, ch := range traversals {
					for range ch {
					}
				}
				return s
			}
		}
	}
}

func traverse(e expr.Expr) <-chan expr.Expr {
	ch := make(chan expr.Expr)
	go func() {
		traverseHelper(ch, e)
		close(ch)
	}()
	return ch
}

func traverseHelper(ch chan<- expr.Expr, e expr.Expr) {
	switch e := e.(type) {
	case *expr.Var:
		ch <- e
	case *expr.Appl:
		traverseHelper(ch, e.Right)
		traverseHelper(ch, e.Left)
	case *expr.Abst:
		traverseHelper(ch, e.Bound)
		traverseHelper(ch, e.Body)
	}
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
			err = &NoMatchError{e.SourceInfo(), e.TypeInfo(), results}
			results = nil
			return
		}
		if len(filtered) > 1 {
			err = &AmbiguousError{e.SourceInfo(), e.TypeInfo(), results}
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
		cannotApplyErr := &CannotApplyError{
			LeftSourceInfo:  e.Left.SourceInfo(),
			RightSourceInfo: e.Right.SourceInfo(),
		}
		for _, r1 := range results1 {
			results2, err := infer(
				varIndex,
				global,
				r1.Subst.ApplyToVars(local),
				e.Right,
			)
			if err != nil {
				cannotApplyErr.Cases = append(cannotApplyErr.Cases, struct {
					Left  types.Type
					Right types.Type
					Err   error
				}{r1.Type, nil, err})
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
						Left  types.Type
						Right types.Type
						Err   error
					}{r1.Type, r2.Type, nil})
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
		var (
			bindType types.Type
			bodyType types.Type
		)
		if f, ok := e.TypeInfo().(*types.Func); ok {
			bindType = f.From
			bodyType = e.Body.TypeInfo()
			if bodyType == nil {
				bodyType = f.To
			}
		} else {
			bindType = newVar(varIndex)
			bodyType = e.Body.TypeInfo()
		}
		newLocal := local.Assume(e.Bound.Name, bindType)
		bodyResults, err := infer(varIndex, global, newLocal, e.Body.WithTypeInfo(bodyType))
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
