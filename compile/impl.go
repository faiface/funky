package compile

import (
	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/parse/parseinfo"
	"github.com/faiface/funky/types"
)

type impl interface {
	SourceInfo() *parseinfo.Source
	TypeInfo() types.Type
}

type (
	implUndefined struct {
		SI   *parseinfo.Source
		Type types.Type
	}
	implExpr struct{ Expr expr.Expr }
)

func (u *implUndefined) SourceInfo() *parseinfo.Source { return u.SI }
func (e *implExpr) SourceInfo() *parseinfo.Source      { return e.Expr.SourceInfo() }

func (u *implUndefined) TypeInfo() types.Type { return u.Type }
func (e *implExpr) TypeInfo() types.Type      { return e.Expr.TypeInfo() }
