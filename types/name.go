package types

import (
	"go/types"

	"github.com/faiface/funky/parse/parseinfo"
)

type Name interface {
	SourceInfo() *parseinfo.Source
	Arity() int
}

type (
	Record struct {
		SI     *parseinfo.Source
		Args   []string
		Fields []struct {
			SI   *parseinfo.Source
			Name string
			Type types.Type
		}
	}

	Enum struct {
		SI           *parseinfo.Source
		Args         []string
		Alternatives []struct {
			SI     *parseinfo.Source
			Name   string
			Fields []types.Type
		}
	}

	Alias struct {
		SI   *parseinfo.Source
		Args []string
		Type types.Type
	}
)

func (r *Record) SourceInfo() *parseinfo.Source { return r.SI }
func (e *Enum) SourceInfo() *parseinfo.Source   { return e.SI }
func (a *Alias) SourceInfo() *parseinfo.Source  { return a.SI }

func (r *Record) Arity() int { return len(r.Args) }
func (e *Enum) Arity() int   { return len(e.Args) }
func (a *Alias) Arity() int  { return len(a.Args) }
