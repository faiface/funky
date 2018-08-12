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
			Left types.Type
			Err  error
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

	CannotSwitchError struct {
		ExprSourceInfo *parseinfo.Source
		Cases          []struct {
			Expr types.Type
			Err  error
		}
	}
)

func (err *CannotApplyError) AddCase(left types.Type, er error) {
	err.Cases = append(err.Cases, struct {
		Left types.Type
		Err  error
	}{left, er})
}

func (err *CannotSwitchError) AddCase(exp types.Type, er error) {
	err.Cases = append(err.Cases, struct {
		Expr types.Type
		Err  error
	}{exp, er})
}

func (err *NotBoundError) Error() string {
	return fmt.Sprintf("%v: variable not bound: %s", err.SourceInfo, err.Name)
}

func (err *CannotApplyError) Error() string {
	s := fmt.Sprintf("%v: cannot apply; in case function has type:", err.LeftSourceInfo)
	for _, cas := range err.Cases {
		s += "\n" + cas.Left.String()
		s += "\n" + indent(cas.Err.Error())
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
	s := fmt.Sprintf("%v: does not match required type: %v\n", err.SourceInfo, err.TypeInfo)
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

func (err *CannotSwitchError) Error() string {
	s := fmt.Sprintf("%v: cannot switch; in case switched expression has type:", err.ExprSourceInfo)
	for _, cas := range err.Cases {
		s += "\n" + cas.Expr.String()
		s += "\n" + indent(cas.Err.Error())
	}
	return s
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
		ch <- e.Bound
		traverseHelper(ch, e.Body)
	case *expr.Switch:
		traverseHelper(ch, e.Expr)
		for i := len(e.Cases) - 1; i >= 0; i-- {
			traverseHelper(ch, e.Cases[i].Body)
		}
	}
}

type InferResult struct {
	Type  types.Type
	Subst Subst
	Expr  expr.Expr
}

func Infer(names map[string]types.Name, global map[string][]types.Type, e expr.Expr) ([]InferResult, error) {
	varIndex := 0
	e = instExpr(&varIndex, e)
	results, err := infer(&varIndex, names, global, make(map[string]types.Type), e)
	if err != nil {
		return nil, err
	}
	for i := range results {
		results[i].Expr = results[i].Subst.ApplyToExpr(results[i].Expr)
	}
	return results, nil
}

var counter = 0

func infer(
	varIndex *int,
	names map[string]types.Name,
	global map[string][]types.Type,
	local map[string]types.Type,
	e expr.Expr,
) (results []InferResult, err error) {
	defer func() {
		if err != nil || e.TypeInfo() == nil {
			return
		}
		// filter infer results by the type info
		var filtered []InferResult
		for _, r := range results {
			if IsSpec(names, r.Type, e.TypeInfo()) {
				s, _ := Unify(names, r.Type, e.TypeInfo())
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
	}()

	switch e := e.(type) {
	case *expr.Char, *expr.Int, *expr.Float:
		return []InferResult{{
			Type:  e.TypeInfo(),
			Subst: nil,
			Expr:  e,
		}}, nil

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
		resultsL, err := infer(varIndex, names, global, local, e.Left)
		if err != nil {
			return nil, err
		}
		resultsR, err := infer(varIndex, names, global, local, e.Right)
		if err != nil {
			return nil, err
		}

		results = nil
		resultType := newVar(varIndex)
		for _, rL := range resultsL {
			for _, rR := range resultsR {
				s, ok := rL.Subst.Unify(names, rR.Subst)
				if !ok {
					continue
				}
				st, ok := Unify(names, s.ApplyToType(rL.Type), &types.Func{
					From: s.ApplyToType(rR.Type),
					To:   resultType,
				})
				if !ok {
					continue
				}
				s = s.Compose(st)
				t := s.ApplyToType(resultType)
				results = append(results, InferResult{
					Type:  t,
					Subst: s,
					Expr: &expr.Appl{
						TI:    t,
						Left:  rL.Expr,
						Right: rR.Expr,
					},
				})
			}
		}

		if len(results) == 0 {
			return nil, fmt.Errorf("%v: type-checking error", e.Right.SourceInfo())
		}
		return results, nil

	case *expr.Abst:
		var (
			bindType = e.Bound.TypeInfo()
			bodyType = e.Body.TypeInfo()
		)
		if f, ok := e.TypeInfo().(*types.Func); ok {
			if bindType == nil {
				bindType = f.From
			}
			if bodyType == nil {
				bodyType = f.To
			}
		} else if bindType == nil {
			bindType = newVar(varIndex)
		}
		newLocal := assume(local, e.Bound.Name, bindType)
		bodyResults, err := infer(varIndex, names, global, newLocal, e.Body.WithTypeInfo(bodyType))
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

	case *expr.Switch:
		exprResults, err := infer(varIndex, names, global, local, e.Expr)
		if err != nil {
			return nil, err
		}

		// check if cases don't produce errors on their own, if so, report simple error
		for i := range e.Cases {
			_, err := infer(varIndex, names, global, local, e.Cases[i].Body)
			if err != nil {
				return nil, err
			}
		}

		results = nil
		cannotSwitchError := &CannotSwitchError{
			ExprSourceInfo: e.Expr.SourceInfo(),
		}

	exprResultsLoop:
		for _, exprResult := range exprResults {
			appl, ok := exprResult.Type.(*types.Appl)
			if !ok {
				cannotSwitchError.AddCase(
					exprResult.Type,
					fmt.Errorf("not a union or too general type; fix your code or add a type annotation"),
				)
				continue
			}
			for {
				alias, ok := names[appl.Name].(*types.Alias)
				if !ok {
					break
				}
				revealed := revealAlias(alias, appl.Args)
				revealedAppl, ok := revealed.(*types.Appl)
				if !ok {
					break
				}
				appl = revealedAppl
			}
			union, ok := names[appl.Name].(*types.Union)
			if !ok {
				cannotSwitchError.AddCase(
					exprResult.Type,
					fmt.Errorf("not a union or too general type; fix your code or add a type annotation"),
				)
				continue
			}

			if len(union.Alts) != len(e.Cases) {
				cannotSwitchError.AddCase(
					exprResult.Type,
					fmt.Errorf(
						"wrong number of cases; %d required, %d given",
						len(union.Alts), len(e.Cases),
					),
				)
				continue
			}
			for i := range union.Alts {
				if union.Alts[i].Name != e.Cases[i].Alt {
					cannotSwitchError.AddCase(
						exprResult.Type,
						fmt.Errorf(
							"wrong case alternative; should be %s, is %s",
							union.Alts[i].Name, e.Cases[i].Alt,
						),
					)
					continue exprResultsLoop
				}
			}

			// build a substitution for the union type arguments
			unionSubst := Subst(nil)
			for i := range union.Args {
				unionSubst = unionSubst.Compose(Subst{union.Args[i]: appl.Args[i]})
			}

			s := exprResult.Subst

			//TODO: use type info (was causing problems, though)
			/*var resultType types.Type
			if e.TypeInfo() != nil {
				resultType = e.TypeInfo()
			} else {
				resultType = newVar(varIndex)
			}*/
			resultType := newVar(varIndex)

			tmpResults := []InferResult{{ // results for this specific inference of Switch.Expr
				Type:  resultType,
				Subst: s,
				Expr: (&expr.Switch{
					TI:    e.TI,
					SI:    e.SI,
					Expr:  exprResult.Expr,
					Cases: e.Cases,
				}).WithTypeInfo(resultType),
			}}

			for i := range e.Cases {
				var newTmpResults []InferResult
				tmpCannotSwitchError := &CannotSwitchError{}

				for _, tmpResult := range tmpResults {
					s := s.Compose(tmpResult.Subst)

					caseType := tmpResult.Type
					for j := len(union.Alts[i].Fields) - 1; j >= 0; j-- {
						caseType = &types.Func{
							From: unionSubst.Compose(s).ApplyToType(union.Alts[i].Fields[j]),
							To:   caseType,
						}
					}

					bodyResults, err := infer(
						varIndex,
						names,
						global,
						s.ApplyToVars(local),
						e.Cases[i].Body,
					)
					if err != nil {
						tmpCannotSwitchError.AddCase(s.ApplyToType(exprResult.Type), err)
						continue
					}

					for _, bodyResult := range bodyResults {
						s := s.Compose(bodyResult.Subst)

						s1, ok := Unify(names, bodyResult.Type, s.ApplyToType(caseType))
						if !ok {
							tmpCannotSwitchError.AddCase(
								s.ApplyToType(exprResult.Type),
								fmt.Errorf(
									"%v: case body type: %v; does not match: %v",
									e.Cases[i].SI,
									bodyResult.Type,
									s.ApplyToType(caseType),
								),
							)
							continue
						}

						s = s.Compose(s1)
						result := InferResult{
							Type:  s.ApplyToType(tmpResult.Type),
							Subst: s,
						}
						result.Expr = tmpResult.Expr.WithTypeInfo(result.Type)
						result.Expr.(*expr.Switch).Cases[i].Body = bodyResult.Expr

						newTmpResults = append(newTmpResults, result)
					}
				}

				if len(newTmpResults) == 0 {
					cannotSwitchError.Cases = append(cannotSwitchError.Cases, tmpCannotSwitchError.Cases...)
				}
				tmpResults = newTmpResults
			}

			results = append(results, tmpResults...)
		}

		if len(results) == 0 {
			return nil, cannotSwitchError
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

func assume(vars map[string]types.Type, v string, t types.Type) map[string]types.Type {
	newVars := make(map[string]types.Type)
	for v, t := range vars {
		newVars[v] = t
	}
	newVars[v] = t
	return newVars
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
