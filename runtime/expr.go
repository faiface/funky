package runtime

import "math/big"

type ctx struct {
	Arg  Expr
	Next *ctx
}

func cons(arg Expr, next *ctx) *ctx {
	return &ctx{
		Arg:  arg,
		Next: next,
	}
}

func drop(n int, ctx *ctx) *ctx {
	for n > 0 {
		ctx = ctx.Next
		n--
	}
	return ctx
}

type Expr interface {
	Reduce() Expr
	WithCtx(*ctx) Expr
	Apply(Expr) Expr
}

type (
	Char  rune
	Int   big.Int
	Float float64

	Record struct {
		Fields []Expr
	}
	Union struct {
		Alternative int
		Fields      []Expr
	}

	Var struct {
		ctx *ctx
	}
	Appl struct {
		reduced      bool
		ctx          *ctx
		LDrop, RDrop int
		Left, Right  Expr
	}
	Abst struct {
		ctx  *ctx
		Body Expr
	}

	Ref struct {
		Ref *Expr
	}
)

func (c Char) Reduce() Expr   { return c }
func (i *Int) Reduce() Expr   { return i }
func (f Float) Reduce() Expr  { return f }
func (r Record) Reduce() Expr { return r }
func (u Union) Reduce() Expr  { return u }
func (v Var) Reduce() Expr    { return v.ctx.Arg }
func (t *Appl) Reduce() Expr {
	if t.reduced {
		return t.Left
	}
	t.Left = t.Left.WithCtx(drop(t.LDrop, t.ctx)).Reduce()
	t.Left = t.Left.Apply(t.Right.WithCtx(drop(t.RDrop, t.ctx))).Reduce()
	t.Right = nil
	return t.Left
}
func (a *Abst) Reduce() Expr { return a }
func (r Ref) Reduce() Expr   { return (*r.Ref).Reduce() }

func (c Char) WithCtx(*ctx) Expr  { return c }
func (i *Int) WithCtx(*ctx) Expr  { return i }
func (f Float) WithCtx(*ctx) Expr { return f }
func (r Record) WithCtx(ctx *ctx) Expr {
	fields := make([]Expr, len(r.Fields))
	for i := range fields {
		fields[i] = r.Fields[i].WithCtx(ctx)
	}
	return Record{Fields: fields}
}
func (u Union) WithCtx(ctx *ctx) Expr {
	fields := make([]Expr, len(u.Fields))
	for i := range fields {
		fields[i] = u.Fields[i].WithCtx(ctx)
	}
	return Union{Alternative: u.Alternative, Fields: fields}
}
func (v Var) WithCtx(ctx *ctx) Expr { return Var{ctx: ctx} }
func (t *Appl) WithCtx(ctx *ctx) Expr {
	return &Appl{
		reduced: t.reduced,
		ctx:     ctx,
		LDrop:   t.LDrop,
		RDrop:   t.RDrop,
		Left:    t.Left,
		Right:   t.Right,
	}
}
func (a *Abst) WithCtx(ctx *ctx) Expr { return &Abst{ctx: ctx, Body: a.Body} }
func (r Ref) WithCtx(ctx *ctx) Expr   { return (*r.Ref).WithCtx(ctx) }

func (c Char) Apply(Expr) Expr      { panic("not applicable") }
func (i *Int) Apply(Expr) Expr      { panic("not applicable") }
func (f Float) Apply(Expr) Expr     { panic("not applicable") }
func (r Record) Apply(Expr) Expr    { panic("not applicable") }
func (u Union) Apply(Expr) Expr     { panic("not applicable") }
func (v Var) Apply(Expr) Expr       { panic("not applicable") }
func (t *Appl) Apply(Expr) Expr     { panic("not applicable") }
func (a *Abst) Apply(arg Expr) Expr { return a.Body.WithCtx(cons(arg, a.ctx)) }
func (r Ref) Apply(arg Expr) Expr   { return (*r.Ref).Apply(arg) }
