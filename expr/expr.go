package expr

import (
	"math/big"

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
	Char struct {
		SI    *parseinfo.Source
		Value rune
	}

	Int struct {
		SI    *parseinfo.Source
		Value *big.Int
	}

	Float struct {
		SI    *parseinfo.Source
		Value float64
	}

	Var struct {
		TI   types.Type
		SI   *parseinfo.Source
		Name string
	}

	Abst struct {
		TI    types.Type
		SI    *parseinfo.Source
		Bound *Var
		Body  Expr
	}

	Appl struct {
		TI    types.Type
		Left  Expr
		Right Expr
	}

	Switch struct {
		TI    types.Type
		SI    *parseinfo.Source
		Expr  Expr
		Cases []struct {
			SI   *parseinfo.Source
			Alt  string
			Body Expr
		}
	}
)

func (c *Char) TypeInfo() types.Type   { return &types.Appl{Name: "Char"} }
func (i *Int) TypeInfo() types.Type    { return &types.Appl{Name: "Int"} }
func (f *Float) TypeInfo() types.Type  { return &types.Appl{Name: "Float"} }
func (v *Var) TypeInfo() types.Type    { return v.TI }
func (a *Abst) TypeInfo() types.Type   { return a.TI }
func (a *Appl) TypeInfo() types.Type   { return a.TI }
func (s *Switch) TypeInfo() types.Type { return s.TI }

func (c *Char) WithTypeInfo(types.Type) Expr   { return c }
func (i *Int) WithTypeInfo(types.Type) Expr    { return i }
func (f *Float) WithTypeInfo(types.Type) Expr  { return f }
func (v *Var) WithTypeInfo(t types.Type) Expr  { return &Var{t, v.SI, v.Name} }
func (a *Abst) WithTypeInfo(t types.Type) Expr { return &Abst{t, a.SI, a.Bound, a.Body} }
func (a *Appl) WithTypeInfo(t types.Type) Expr { return &Appl{t, a.Left, a.Right} }
func (s *Switch) WithTypeInfo(t types.Type) Expr {
	newCases := make([]struct {
		SI   *parseinfo.Source
		Alt  string
		Body Expr
	}, len(s.Cases))
	copy(newCases, s.Cases)
	return &Switch{t, s.SI, s.Expr, newCases}
}

func (c *Char) SourceInfo() *parseinfo.Source   { return c.SI }
func (i *Int) SourceInfo() *parseinfo.Source    { return i.SI }
func (f *Float) SourceInfo() *parseinfo.Source  { return f.SI }
func (v *Var) SourceInfo() *parseinfo.Source    { return v.SI }
func (a *Abst) SourceInfo() *parseinfo.Source   { return a.SI }
func (a *Appl) SourceInfo() *parseinfo.Source   { return a.Left.SourceInfo() }
func (s *Switch) SourceInfo() *parseinfo.Source { return s.SI }

func (c *Char) Map(f func(Expr) Expr) Expr   { return f(c) }
func (i *Int) Map(f func(Expr) Expr) Expr    { return f(i) }
func (f *Float) Map(fn func(Expr) Expr) Expr { return fn(f) }
func (v *Var) Map(f func(Expr) Expr) Expr    { return f(v) }
func (a *Abst) Map(f func(Expr) Expr) Expr {
	return f(&Abst{a.TI, a.SI, a.Bound.Map(f).(*Var), a.Body.Map(f)})
}
func (a *Appl) Map(f func(Expr) Expr) Expr { return f(&Appl{a.TI, a.Left.Map(f), a.Right.Map(f)}) }
func (s *Switch) Map(f func(Expr) Expr) Expr {
	newCases := make([]struct {
		SI   *parseinfo.Source
		Alt  string
		Body Expr
	}, len(s.Cases))
	for i := range newCases {
		newCases[i].SI = s.Cases[i].SI
		newCases[i].Alt = s.Cases[i].Alt
		newCases[i].Body = s.Cases[i].Body.Map(f)
	}
	return f(&Switch{s.TI, s.SI, s.Expr.Map(f), newCases})
}
