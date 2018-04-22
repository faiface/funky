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
	reduce() Expr
	withCtx(*ctx) Expr
	apply(Expr) Expr
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
		ctx   *ctx
		Index int
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
		Expr *Expr
	}

	GoFunc func(Expr) Expr
)

func (c Char) reduce() Expr   { return c }
func (i *Int) reduce() Expr   { return i }
func (f Float) reduce() Expr  { return f }
func (r Record) reduce() Expr { return r }
func (u Union) reduce() Expr  { return u }
func (v Var) reduce() Expr    { return drop(v.Index, v.ctx).Arg }
func (t *Appl) reduce() Expr {
	if t.reduced {
		return t.Left
	}
	t.Left = t.Left.withCtx(drop(t.LDrop, t.ctx)).reduce()
	t.Left = t.Left.apply(t.Right.withCtx(drop(t.RDrop, t.ctx))).reduce()
	t.Right = nil
	t.reduced = true
	return t.Left
}
func (a *Abst) reduce() Expr  { return a }
func (r Ref) reduce() Expr    { return (*r.Expr).reduce() }
func (g GoFunc) reduce() Expr { return g }

func (c Char) withCtx(*ctx) Expr  { return c }
func (i *Int) withCtx(*ctx) Expr  { return i }
func (f Float) withCtx(*ctx) Expr { return f }
func (r Record) withCtx(ctx *ctx) Expr {
	fields := make([]Expr, len(r.Fields))
	for i := range fields {
		fields[i] = r.Fields[i].withCtx(ctx)
	}
	return Record{Fields: fields}
}
func (u Union) withCtx(ctx *ctx) Expr {
	fields := make([]Expr, len(u.Fields))
	for i := range fields {
		fields[i] = u.Fields[i].withCtx(ctx)
	}
	return Union{Alternative: u.Alternative, Fields: fields}
}
func (v Var) withCtx(ctx *ctx) Expr { return Var{ctx: ctx, Index: v.Index} }
func (t *Appl) withCtx(ctx *ctx) Expr {
	return &Appl{
		reduced: t.reduced,
		ctx:     ctx,
		LDrop:   t.LDrop,
		RDrop:   t.RDrop,
		Left:    t.Left,
		Right:   t.Right,
	}
}
func (a *Abst) withCtx(ctx *ctx) Expr { return &Abst{ctx: ctx, Body: a.Body} }
func (r Ref) withCtx(ctx *ctx) Expr   { return (*r.Expr).withCtx(ctx) }
func (g GoFunc) withCtx(*ctx) Expr    { return g }

func (c Char) apply(Expr) Expr       { panic("not applicable") }
func (i *Int) apply(Expr) Expr       { panic("not applicable") }
func (f Float) apply(Expr) Expr      { panic("not applicable") }
func (r Record) apply(Expr) Expr     { panic("not applicable") }
func (u Union) apply(Expr) Expr      { panic("not applicable") }
func (v Var) apply(Expr) Expr        { panic("not applicable") }
func (t *Appl) apply(Expr) Expr      { panic("not applicable") }
func (a *Abst) apply(arg Expr) Expr  { return a.Body.withCtx(cons(arg, a.ctx)) }
func (r Ref) apply(arg Expr) Expr    { return (*r.Expr).apply(arg) }
func (g GoFunc) apply(arg Expr) Expr { return g(arg) }
