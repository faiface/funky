package expr

import (
	"fmt"
	"strconv"
)

func (v *Var) leftString() string    { return v.Name }
func (a *Appl) leftString() string   { return a.String() }
func (a *Abst) leftString() string   { return "(" + a.String() + ")" }
func (s *Switch) leftString() string { return "(" + s.String() + ")" }
func (c *Char) leftString() string   { return c.String() }
func (i *Int) leftString() string    { return i.String() }
func (f *Float) leftString() string  { return f.String() }

func (v *Var) rightString() string    { return v.Name }
func (a *Appl) rightString() string   { return "(" + a.String() + ")" }
func (a *Abst) rightString() string   { return "(" + a.String() + ")" }
func (s *Switch) rightString() string { return s.String() }
func (c *Char) rightString() string   { return c.String() }
func (i *Int) rightString() string    { return i.String() }
func (f *Float) rightString() string  { return f.String() }

func (v *Var) String() string { return v.Name }
func (a *Appl) String() string {
	return fmt.Sprintf("%s %s", a.Left.leftString(), a.Right.rightString())
}
func (a *Abst) String() string { return fmt.Sprintf("\\%v %v", a.Bound, a.Body) }
func (s *Switch) String() string {
	str := fmt.Sprintf("switch %v", s.Expr.String())
	for _, cas := range s.Cases {
		str += fmt.Sprintf(" case %s %v", cas.Alt, cas.Body.String())
	}
	return str
}
func (c *Char) String() string  { return strconv.QuoteRune(c.Value) }
func (i *Int) String() string   { return i.Value.Text(10) }
func (f *Float) String() string { return fmt.Sprint(f.Value) }
