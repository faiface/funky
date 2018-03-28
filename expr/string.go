package expr

import "fmt"

//TODO: print type info (?)

func (v *Var) leftString() string { return v.Name }
func (a *Appl) leftString() string {
	return fmt.Sprintf("%s %s", a.Left.leftString(), a.Right.rightString())
}
func (a *Abst) leftString() string { return fmt.Sprintf("(λ%v %v)", a.Bound, a.Body) }

func (v *Var) rightString() string { return v.Name }
func (a *Appl) rightString() string {
	return fmt.Sprintf("(%s %s)", a.Left.leftString(), a.Right.rightString())
}
func (a *Abst) rightString() string { return fmt.Sprintf("(λ%v %v)", a.Bound, a.Body) }

func (v *Var) String() string  { return v.Name }
func (a *Appl) String() string { return a.leftString() }
func (a *Abst) String() string { return fmt.Sprintf("λ%v %v", a.Bound, a.Body) }
