package expr

import "github.com/faiface/funky/parse/parseinfo"

type Expr interface {
	leftString() string
	rightString() string
	String() string

	SourceInfo() *parseinfo.Source
}

type (
	Var struct {
		SI   *parseinfo.Source
		Name string
	}

	Appl struct {
		Left, Right Expr
	}

	Abst struct {
		Bound *Var
		Body  Expr
	}
)

func (v *Var) SourceInfo() *parseinfo.Source  { return v.SI }
func (a *Appl) SourceInfo() *parseinfo.Source { return a.Left.SourceInfo() }
func (a *Abst) SourceInfo() *parseinfo.Source { return a.Bound.SourceInfo() }
