package runtime

import "math/big"

type State interface {
	Reduce() State
}

type (
	Char  struct{ Value rune }
	Int   struct{ Value *big.Int }
	Float struct{ Value float64 }

	Record struct {
		Fields []State
	}

	Union struct {
		Alternative int
		Fields      []State
	}

	GoFunc func(*Data) State
)

func (c *Char) Reduce() State   { return c }
func (i *Int) Reduce() State    { return i }
func (f *Float) Reduce() State  { return f }
func (r *Record) Reduce() State { return r }
func (u *Union) Reduce() State  { return u }
func (gf GoFunc) Reduce() State { return gf }

func (r *Record) Field(i int) State {
	r.Fields[i] = r.Fields[i].Reduce()
	return r.Fields[i]
}

func (u *Union) Field(i int) State {
	u.Fields[i] = u.Fields[i].Reduce()
	return u.Fields[i]
}
