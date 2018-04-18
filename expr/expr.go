package expr

import (
	"github.com/faiface/funky/parse/parseinfo"
	"github.com/faiface/funky/types"
)

type Expr interface {
	leftString() string
	rightString() string
	String() string

	TypeInfo() types.Type
	WithTypeInfo(types.Type) Expr
	SourceInfo() *parseinfo.Source

	Map(func(Expr) Expr) Expr
}

type (
	Var struct {
		TI   types.Type
		SI   *parseinfo.Source
		Name string
	}

	Appl struct {
		TI    types.Type
		Left  Expr
		Right Expr
	}

	Abst struct {
		TI    types.Type
		SI    *parseinfo.Source
		Bound *Var
		Body  Expr
	}

	Switch struct {
		TI    types.Type
		SI    *parseinfo.Source
		Expr  Expr
		Cases []struct {
			Alt  *Var
			Body Expr
		}
	}
)

func (v *Var) TypeInfo() types.Type    { return v.TI }
func (a *Appl) TypeInfo() types.Type   { return a.TI }
func (a *Abst) TypeInfo() types.Type   { return a.TI }
func (s *Switch) TypeInfo() types.Type { return s.TI }

func (v *Var) WithTypeInfo(t types.Type) Expr    { return &Var{t, v.SI, v.Name} }
func (a *Appl) WithTypeInfo(t types.Type) Expr   { return &Appl{t, a.Left, a.Right} }
func (a *Abst) WithTypeInfo(t types.Type) Expr   { return &Abst{t, a.SI, a.Bound, a.Body} }
func (s *Switch) WithTypeInfo(t types.Type) Expr { return &Switch{t, s.SI, s.Expr, s.Cases} }

func (v *Var) SourceInfo() *parseinfo.Source    { return v.SI }
func (a *Appl) SourceInfo() *parseinfo.Source   { return a.Left.SourceInfo() }
func (a *Abst) SourceInfo() *parseinfo.Source   { return a.SI }
func (s *Switch) SourceInfo() *parseinfo.Source { return s.SI }

func (v *Var) Map(f func(Expr) Expr) Expr  { return f(v) }
func (a *Appl) Map(f func(Expr) Expr) Expr { return f(&Appl{a.TI, a.Left.Map(f), a.Right.Map(f)}) }
func (a *Abst) Map(f func(Expr) Expr) Expr {
	return f(&Abst{a.TI, a.SI, a.Bound.Map(f).(*Var), a.Body.Map(f)})
}
func (s *Switch) Map(f func(Expr) Expr) Expr {
	newCases := make([]struct {
		Alt  *Var
		Body Expr
	}, len(s.Cases))
	for i := range newCases {
		newCases[i].Alt = s.Cases[i].Alt.Map(f).(*Var)
		newCases[i].Body = s.Cases[i].Body.Map(f)
	}
	return f(&Switch{s.TI, s.SI, s.Expr.Map(f), newCases})
}
