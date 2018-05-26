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
