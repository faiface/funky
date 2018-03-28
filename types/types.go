package types

import "github.com/faiface/funky/parse/parseinfo"

type Type interface {
	leftString() string
	insideString() string
	String() string

	SourceInfo() *parseinfo.Source
}

type (
	Var struct {
		SI   *parseinfo.Source
		Name string
	}

	Appl struct {
		Cons *Var // type constructor (e.g. List, Map, Int, ...)
		Args []Type
	}

	Func struct {
		From, To Type
	}
)

type Scheme struct {
	Bound []string
	Body  Type
}

func (v *Var) SourceInfo() *parseinfo.Source  { return v.SI }
func (a *Appl) SourceInfo() *parseinfo.Source { return a.Cons.SourceInfo() }
func (f *Func) SourceInfo() *parseinfo.Source { return f.From.SourceInfo() }
