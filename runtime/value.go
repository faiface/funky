package runtime

import (
	"math/big"
	"strings"
)

type Value struct {
	State State
}

func (v *Value) reduce() { v.State = v.State.Reduce() }

func (v *Value) Char() rune     { v.reduce(); return v.State.(*Char).Value }
func (v *Value) Int() *big.Int  { v.reduce(); return v.State.(*Int).Value }
func (v *Value) Float() float64 { v.reduce(); return v.State.(*Float).Value }

func (v *Value) Field(i int) *Value {
	v.reduce()
	switch state := v.State.(type) {
	case *Record:
		return &Value{state.Field(i)}
	case *Union:
		return &Value{state.Field(i)}
	default:
		panic("not a record or a union")
	}
}

func (v *Value) Alternative() int {
	v.reduce()
	return v.State.(*Union).Alternative
}

func (v *Value) Apply(arg *Value) *Value {
	v.reduce()
	thunk := v.State.(*Thunk)
	if thunk.Code.Kind != CodeAbst {
		panic("not an abstraction")
	}
	return &Value{&Thunk{
		thunk.Code.A,
		Cons(arg.State, Drop(thunk.Code.Drop, thunk.Data)),
		nil,
	}}
}

func (v *Value) Bool() bool {
	return v.Alternative() == 0
}

func (v *Value) List() []*Value {
	var list []*Value
	for v.Alternative() != 0 { // empty | (::)
		list = append(list, v.Field(0)) // first
		v = v.Field(1)                  // rest
	}
	return list
}

func (v *Value) String() string {
	var builder strings.Builder
	for v.Alternative() != 0 { // empty | (::)
		builder.WriteRune(v.Field(0).Char()) // first
		v = v.Field(1)                       // rest
	}
	return builder.String()
}

func MkChar(c rune) *Value     { return &Value{&Char{c}} }
func MkInt(i *big.Int) *Value  { return &Value{&Int{i}} }
func MkFloat(f float64) *Value { return &Value{&Float{f}} }

func MkRecord(fields ...*Value) *Value {
	record := &Record{Fields: make([]State, len(fields))}
	for i := range record.Fields {
		record.Fields[i] = fields[i].State
	}
	return &Value{record}
}

func MkUnion(alternative int, fields ...*Value) *Value {
	union := &Union{Alternative: alternative, Fields: make([]State, len(fields))}
	for i := range union.Fields {
		union.Fields[i] = fields[i].State
	}
	return &Value{union}
}

func MkBool(b bool) *Value {
	if b {
		return MkUnion(0, nil)
	}
	return MkUnion(1, nil)
}

func MkList(elems ...*Value) *Value {
	list := &Union{Alternative: 0}
	for i := len(elems) - 1; i >= 0; i-- {
		list = &Union{Alternative: 1, Fields: []State{elems[i].State, list}}
	}
	return &Value{list}
}

func MkString(s string) *Value {
	str := &Union{Alternative: 0}
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		str = &Union{Alternative: 1, Fields: []State{&Char{runes[i]}, str}}
	}
	return &Value{str}
}
