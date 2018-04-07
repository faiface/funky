package types

import "github.com/faiface/funky/parse/parseinfo"

type Type interface {
	leftString() string
	insideString() string
	String() string

	SourceInfo() *parseinfo.Source

	Map(func(Type) Type) Type
}

type (
	Var struct {
		SI   *parseinfo.Source
		Name string
	}

	Appl struct {
		SI   *parseinfo.Source
		Cons string // type constructor (e.g. List, Map, Int, ...)
		Args []Type
	}

	Func struct {
		From, To Type
	}
)

func (v *Var) SourceInfo() *parseinfo.Source  { return v.SI }
func (a *Appl) SourceInfo() *parseinfo.Source { return a.SI }
func (f *Func) SourceInfo() *parseinfo.Source { return f.From.SourceInfo() }

func (v *Var) Map(f func(Type) Type) Type { return f(v) }
func (a *Appl) Map(f func(Type) Type) Type {
	mapped := &Appl{
		SI:   a.SI,
		Cons: a.Cons,
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
