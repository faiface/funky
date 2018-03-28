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
	SetTypeInfo(types.Type)
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

func (v *Var) SetTypeInfo(t types.Type)  { v.TI = t }
func (a *Appl) SetTypeInfo(t types.Type) { a.TI = t }
func (a *Abst) SetTypeInfo(t types.Type) { a.TI = t }

func (v *Var) SourceInfo() *parseinfo.Source  { return v.SI }
func (a *Appl) SourceInfo() *parseinfo.Source { return a.Left.SourceInfo() }
func (a *Abst) SourceInfo() *parseinfo.Source { return a.Bound.SourceInfo() }
