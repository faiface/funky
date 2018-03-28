package expr

import (
	"go/types"

	"github.com/faiface/funky/parse/parseinfo"
)

type Expr interface {
	leftString() string
	rightString() string
	String() string

	TypeInfo() types.Type
	SourceInfo() *parseinfo.Source
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
		Bound *Var
		Body  Expr
	}
)

func (v *Var) TypeInfo() types.Type  { return v.TI }
func (a *Appl) TypeInfo() types.Type { return a.TI }
func (a *Abst) TypeInfo() types.Type { return a.TI }

func (v *Var) SourceInfo() *parseinfo.Source  { return v.SI }
func (a *Appl) SourceInfo() *parseinfo.Source { return a.Left.SourceInfo() }
func (a *Abst) SourceInfo() *parseinfo.Source { return a.Bound.SourceInfo() }
