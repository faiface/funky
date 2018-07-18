package runtime

import "math/big"

type Value interface {
	Reduce() Value
}

type (
	Char  struct{ Value rune }
	Int   struct{ Value *big.Int }
	Float struct{ Value float64 }

	Record struct {
		Fields []Value
	}

	Union struct {
		Alternative int
		Fields      []Value
	}

	GoFunc func(*Data) Value
)

func (c *Char) Reduce() Value   { return c }
func (i *Int) Reduce() Value    { return i }
func (f *Float) Reduce() Value  { return f }
func (r *Record) Reduce() Value { return r }
func (u *Union) Reduce() Value  { return u }
func (gf GoFunc) Reduce() Value { return gf }

func (r *Record) Field(i int) Value {
	r.Fields[i] = r.Fields[i].Reduce()
	return r.Fields[i]
}

func (u *Union) Field(i int) Value {
	u.Fields[i] = u.Fields[i].Reduce()
	return u.Fields[i]
}
