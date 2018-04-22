package runtime

import (
	"math/big"
	"strings"
)

type Value struct {
	Expr Expr
}

func (v *Value) reduce() { v.Expr = v.Expr.Reduce() }

func (v *Value) Char() rune     { v.reduce(); return rune(v.Expr.(Char)) }
func (v *Value) Int() *big.Int  { v.reduce(); return (*big.Int)(v.Expr.(*Int)) }
func (v *Value) Float() float64 { v.reduce(); return float64(v.Expr.(Float)) }
func (v *Value) Field(i int) *Value {
	v.reduce()
	switch v := v.Expr.(type) {
	case Record:
		return &Value{v.Fields[i]}
	case Union:
		return &Value{v.Fields[i]}
	default:
		panic("not a record or a union")
	}
}
func (v *Value) Alternative() int        { v.reduce(); return v.Expr.(Union).Alternative }
func (v *Value) Apply(arg *Value) *Value { v.reduce(); return &Value{v.Expr.Apply(arg.Expr)} }

func (v *Value) Bool() bool {
	return v.Alternative() == 0 // true | false
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

func MkChar(c rune) *Value     { return &Value{Char(c)} }
func MkInt(i *big.Int) *Value  { return &Value{(*Int)(i)} }
func MkFloat(f float64) *Value { return &Value{Float(f)} }

//TODO: other constructors
