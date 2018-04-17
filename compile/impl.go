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
	implUndefined struct{ typ types.Type }
	implExpr      struct{ expr expr.Expr }
)

func (iu *implUndefined) SourceInfo() *parseinfo.Source { return nil }
func (ie *implExpr) SourceInfo() *parseinfo.Source      { return ie.expr.SourceInfo() }

func (iu *implUndefined) TypeInfo() types.Type { return iu.typ }
func (ie *implExpr) TypeInfo() types.Type      { return ie.expr.TypeInfo() }
