package types

import "fmt"

func (v *Var) leftString() string  { return v.Name }
func (a *Appl) leftString() string { return a.String() }
func (f *Func) leftString() string { return "(" + f.String() + ")" }

func (v *Var) insideString() string { return v.Name }
func (a *Appl) insideString() string {
	if len(a.Args) > 0 {
		return "(" + a.String() + ")"
	}
	return a.String()
}
func (f *Func) insideString() string { return "(" + f.String() + ")" }

func (v *Var) String() string { return v.Name }
func (a *Appl) String() string {
	s := a.Cons
	for _, arg := range a.Args {
		s += " " + arg.insideString()
	}
	return s
}
func (f *Func) String() string {
	return fmt.Sprintf("%v -> %v", f.From.leftString(), f.To.String())
}
