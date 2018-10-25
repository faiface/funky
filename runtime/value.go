package runtime

import (
	"math/big"
	"strings"

	cxr "github.com/faiface/crux/runtime"
)

type Value struct {
	Globals []cxr.Value
	Value   cxr.Value
}

func (v *Value) reduce() { v.Value = cxr.Reduce(v.Globals, v.Value) }

func (v *Value) Char() rune     { v.reduce(); return v.Value.(*cxr.Char).Value }
func (v *Value) Int() *big.Int  { v.reduce(); return &v.Value.(*cxr.Int).Value }
func (v *Value) Float() float64 { v.reduce(); return v.Value.(*cxr.Float).Value }

func (v *Value) Alternative() int {
	v.reduce()
	return int(v.Value.(*cxr.Struct).Index)
}

func (v *Value) Field(i int) *Value {
	v.reduce()
	str := v.Value.(*cxr.Struct)
	index := len(str.Values) - i - 1
	return &Value{v.Globals, str.Values[index]}
}

func (v *Value) Apply(args ...*Value) *Value {
	values := make([]cxr.Value, len(args))
	for i := range values {
		values[i] = args[i].Value
	}
	return &Value{v.Globals, cxr.Reduce(v.Globals, v.Value, values...)}
}

func (v *Value) Bool() bool {
	return v.Alternative() == 0
}

func (v *Value) List() []*Value {
	var list []*Value
	for v.Alternative() != 0 {
		list = append(list, v.Field(0))
		v = v.Field(1)
	}
	return list
}

func (v *Value) String() string {
	var b strings.Builder
	for v.Alternative() != 0 {
		b.WriteRune(v.Field(0).Char())
		v = v.Field(1)
	}
	return b.String()
}

func MkChar(c rune) *Value {
	return &Value{nil, &cxr.Char{Value: c}}
}

func MkInt(i *big.Int) *Value {
	var v cxr.Int
	v.Value.Set(i)
	return &Value{nil, &v}
}

func MkInt64(i int64) *Value {
	var v cxr.Int
	v.Value.SetInt64(i)
	return &Value{nil, &v}
}

func MkFloat(f float64) *Value {
	return &Value{nil, &cxr.Float{Value: f}}
}

func MkRecord(fields ...*Value) *Value {
	str := &cxr.Struct{Index: 0, Values: make([]cxr.Value, 0, len(fields))}
	for i := len(fields) - 1; i >= 0; i-- {
		str.Values = append(str.Values, fields[i].Value)
	}
	return &Value{nil, str}
}

func MkUnion(alternative int, fields ...*Value) *Value {
	str := &cxr.Struct{Index: int32(alternative), Values: make([]cxr.Value, 0, len(fields))}
	for i := len(fields) - 1; i >= 0; i-- {
		str.Values = append(str.Values, fields[i].Value)
	}
	return &Value{nil, str}
}

func MkBool(b bool) *Value {
	index := int32(0)
	if !b {
		index = 1
	}
	return &Value{nil, &cxr.Struct{Index: index}}
}

func MkList(elems ...*Value) *Value {
	list := &cxr.Struct{Index: 0}
	for i := len(elems) - 1; i >= 0; i-- {
		list = &cxr.Struct{Index: 1, Values: []cxr.Value{list, elems[i].Value}}
	}
	return &Value{nil, list}
}

func MkString(s string) *Value {
	str := &cxr.Struct{Index: 0}
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		str = &cxr.Struct{Index: 1, Values: []cxr.Value{str, &cxr.Char{Value: runes[i]}}}
	}
	return &Value{nil, str}
}
