package types

import "github.com/faiface/funky/parse/parseinfo"

type Name interface {
	SourceInfo() *parseinfo.Source
	Arity() int
}

type (
	Record struct {
		SI     *parseinfo.Source
		Args   []string
		Fields []Field
	}

	Enum struct {
		SI   *parseinfo.Source
		Args []string
		Alts []Alternative
	}

	Alias struct {
		SI   *parseinfo.Source
		Args []string
		Type Type
	}
)

type Field struct {
	SI   *parseinfo.Source
	Name string
	Type Type
}

type Alternative struct {
	SI     *parseinfo.Source
	Name   string
	Fields []Type
}

func (r *Record) SourceInfo() *parseinfo.Source { return r.SI }
func (e *Enum) SourceInfo() *parseinfo.Source   { return e.SI }
func (a *Alias) SourceInfo() *parseinfo.Source  { return a.SI }

func (r *Record) Arity() int { return len(r.Args) }
func (e *Enum) Arity() int   { return len(e.Args) }
func (a *Alias) Arity() int  { return len(a.Args) }
