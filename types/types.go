package types

import "github.com/faiface/funky/parse/parseinfo"

type Type interface {
	leftString() string
	insideString() string
	String() string

	SourceInfo() *parseinfo.Source

	Equal(Type) bool
	Map(func(Type) Type) Type
}

type (
	Var struct {
		SI   *parseinfo.Source
		Name string
	}

	Appl struct {
		SI   *parseinfo.Source
		Name string // type name (e.g. List, Map, Int, ...)
		Args []Type
	}

	Func struct {
		From, To Type
	}
)

func (v *Var) SourceInfo() *parseinfo.Source  { return v.SI }
func (a *Appl) SourceInfo() *parseinfo.Source { return a.SI }
func (f *Func) SourceInfo() *parseinfo.Source { return f.From.SourceInfo() }

func (v *Var) Equal(t Type) bool {
	tv, ok := t.(*Var)
	return ok && v.Name == tv.Name
}
func (a *Appl) Equal(t Type) bool {
	ta, ok := t.(*Appl)
	if !ok || a.Name != ta.Name || len(a.Args) != len(ta.Args) {
		return false
	}
	for i := range a.Args {
		if !a.Args[i].Equal(ta.Args[i]) {
			return false
		}
	}
	return true
}
func (f *Func) Equal(t Type) bool {
	tf, ok := t.(*Func)
	return ok && f.From.Equal(tf.From) && f.To.Equal(tf.To)
}

func (v *Var) Map(f func(Type) Type) Type { return f(v) }
func (a *Appl) Map(f func(Type) Type) Type {
	mapped := &Appl{
		SI:   a.SI,
		Name: a.Name,
		Args: make([]Type, len(a.Args)),
	}
	for i := range mapped.Args {
		mapped.Args[i] = a.Args[i].Map(f)
	}
	return f(mapped)
}
func (f *Func) Map(mf func(Type) Type) Type {
	return mf(&Func{
		From: f.From.Map(mf),
		To:   f.To.Map(mf),
	})
}
