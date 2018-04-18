package expr

import "fmt"

func (v *Var) leftString() string    { return v.Name }
func (a *Appl) leftString() string   { return a.String() }
func (a *Abst) leftString() string   { return "(" + a.String() + ")" }
func (s *Switch) leftString() string { return "(" + s.String() + ")" }

func (v *Var) rightString() string    { return v.Name }
func (a *Appl) rightString() string   { return "(" + a.String() + ")" }
func (a *Abst) rightString() string   { return "(" + a.String() + ")" }
func (s *Switch) rightString() string { return s.String() }

func (v *Var) String() string { return v.Name }
func (a *Appl) String() string {
	return fmt.Sprintf("%s %s", a.Left.leftString(), a.Right.rightString())
}
func (a *Abst) String() string { return fmt.Sprintf("λ%v %v", a.Bound, a.Body) }
func (s *Switch) String() string {
	str := fmt.Sprintf("switch %v", s.Expr.String())
	for _, cas := range s.Cases {
		str += fmt.Sprintf(" case %s %v", cas.Alt, cas.Body.String())
	}
	return str
}
