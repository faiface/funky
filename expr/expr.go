package expr

import "fmt"

type SourceInfo struct {
	Filename     string
	Line, Column int
}

func (si *SourceInfo) String() string {
	if si == nil {
		return "<unknown source>"
	}
	return fmt.Sprintf("%s:%d:%d", si.Filename, si.Line, si.Column)
}

type Expr interface {
	leftString() string
	rightString() string
	String() string

	SourceInfo() *SourceInfo
}

type (
	Var struct {
		SI   *SourceInfo
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

func (v *Var) SourceInfo() *SourceInfo  { return v.SI }
func (a *Appl) SourceInfo() *SourceInfo { return a.Left.SourceInfo() }
func (a *Abst) SourceInfo() *SourceInfo { return a.Bound.SourceInfo() }
