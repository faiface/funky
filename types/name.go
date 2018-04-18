package types

import "github.com/faiface/funky/parse/parseinfo"

type Name interface {
	SourceInfo() *parseinfo.Source
	Arity() int
}

type (
	Builtin struct {
		NumArgs int
	}

	Record struct {
		SI     *parseinfo.Source
		Args   []string
		Fields []Field
	}

	Union struct {
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

func (b *Builtin) SourceInfo() *parseinfo.Source { return nil }
func (r *Record) SourceInfo() *parseinfo.Source  { return r.SI }
func (e *Union) SourceInfo() *parseinfo.Source   { return e.SI }
func (a *Alias) SourceInfo() *parseinfo.Source   { return a.SI }

func (b *Builtin) Arity() int { return b.NumArgs }
func (r *Record) Arity() int  { return len(r.Args) }
func (e *Union) Arity() int   { return len(e.Args) }
func (a *Alias) Arity() int   { return len(a.Args) }
