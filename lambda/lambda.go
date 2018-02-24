package lamdba

import "fmt"

type Node interface {
	fmt.Stringer
}

type Abst struct {
	Bound string
	Body  Node
}

type Appl struct {
	Left  Node
	Right Node
}

type Var struct {
	Name string
}

type Global struct {
	Name string
}

type Eval struct {
	Var  string
	Expr Node
}

type Int64 struct {
	Value int64
}

type Float64 struct {
	Value float64
}

func (a *Abst) String() string    { return fmt.Sprintf("λ%s %v", a.Bound, a.Body) }
func (a *Appl) String() string    { return fmt.Sprintf("(%v %v)", a.Left, a.Right) }
func (v *Var) String() string     { return v.Name }
func (g *Global) String() string  { return g.Name }
func (e *Eval) String() string    { return fmt.Sprintf("(eval %s; %v)", e.Var, e.Expr) }
func (i *Int64) String() string   { return fmt.Sprintf("%d", i.Value) }
func (f *Float64) String() string { return fmt.Sprintf("%f", f.Value) }
