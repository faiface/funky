package runtime

import (
	"math/big"
	"strings"
)

type Box struct {
	Value Value
}

func (v *Box) reduce() { v.Value = v.Value.Reduce() }

func (v *Box) Char() rune     { v.reduce(); return v.Value.(*Char).Value }
func (v *Box) Int() *big.Int  { v.reduce(); return v.Value.(*Int).Value }
func (v *Box) Float() float64 { v.reduce(); return v.Value.(*Float).Value }

func (v *Box) Field(i int) *Box {
	v.reduce()
	switch value := v.Value.(type) {
	case *Record:
		return &Box{value.Field(i)}
	case *Union:
		return &Box{value.Field(i)}
	default:
		panic("not a record or a union")
	}
}

func (v *Box) Alternative() int {
	v.reduce()
	return v.Value.(*Union).Alternative
}

func (v *Box) Apply(arg *Box) *Box {
	v.reduce()
	thunk := v.Value.(*Thunk)
	if thunk.Code.Kind != CodeAbst {
		panic("not an abstraction")
	}
	return &Box{&Thunk{
		thunk.Code.A,
		Cons(arg.Value, Drop(thunk.Code.Drop, thunk.Data)),
		nil,
	}}
}

func (v *Box) Bool() bool {
	return v.Alternative() == 0
}

func (v *Box) List() []*Box {
	var list []*Box
	for v.Alternative() != 0 { // empty | (::)
		list = append(list, v.Field(0)) // first
		v = v.Field(1)                  // rest
	}
	return list
}

func (v *Box) String() string {
	var builder strings.Builder
	for v.Alternative() != 0 { // empty | (::)
		builder.WriteRune(v.Field(0).Char()) // first
		v = v.Field(1)                       // rest
	}
	return builder.String()
}

func MkChar(c rune) *Box     { return &Box{&Char{c}} }
func MkInt(i *big.Int) *Box  { return &Box{&Int{i}} }
func MkFloat(f float64) *Box { return &Box{&Float{f}} }

func MkRecord(fields ...*Box) *Box {
	record := &Record{Fields: make([]Value, len(fields))}
	for i := range record.Fields {
		record.Fields[i] = fields[i].Value
	}
	return &Box{record}
}

func MkUnion(alternative int, fields ...*Box) *Box {
	union := &Union{Alternative: alternative, Fields: make([]Value, len(fields))}
	for i := range union.Fields {
		union.Fields[i] = fields[i].Value
	}
	return &Box{union}
}

func MkBool(b bool) *Box {
	if b {
		return MkUnion(0)
	}
	return MkUnion(1)
}

func MkList(elems ...*Box) *Box {
	list := &Union{Alternative: 0}
	for i := len(elems) - 1; i >= 0; i-- {
		list = &Union{Alternative: 1, Fields: []Value{elems[i].Value, list}}
	}
	return &Box{list}
}

func MkString(s string) *Box {
	str := &Union{Alternative: 0}
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		str = &Union{Alternative: 1, Fields: []Value{&Char{runes[i]}, str}}
	}
	return &Box{str}
}
