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
	Devar() Expr
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
	Switch struct {
		ctx   *ctx
		Expr  Expr
		Cases []Expr
	}

	Ref struct {
		Expr *Expr
	}

	GoFunc func(Expr) Expr
)

func (c Char) Reduce() Expr   { return c }
func (i *Int) Reduce() Expr   { return i }
func (f Float) Reduce() Expr  { return f }
func (r Record) Reduce() Expr { return r }
func (u Union) Reduce() Expr  { return u }
func (v Var) Reduce() Expr    { return drop(v.Index, v.ctx).Arg.Reduce() }
func (a *Appl) Reduce() Expr {
	if a.reduced {
		return a.Left
	}
	a.Left = a.Left.WithCtx(drop(a.LDrop, a.ctx)).Reduce()
	a.Left = a.Left.Apply(a.Right.WithCtx(drop(a.RDrop, a.ctx)).Devar()).Reduce()
	a.Right = nil
	a.reduced = true
	return a.Left
}
func (a *Abst) Reduce() Expr { return a }
func (s Switch) Reduce() Expr {
	union := s.Expr.WithCtx(s.ctx).Reduce().(Union)
	caseExpr := s.Cases[union.Alternative].WithCtx(s.ctx)
	for _, field := range union.Fields {
		caseExpr = caseExpr.Reduce().Apply(field)
	}
	return caseExpr.Reduce()
}
func (r Ref) Reduce() Expr    { return (*r.Expr).Reduce() }
func (g GoFunc) Reduce() Expr { return g }

func (c Char) Devar() Expr   { return c }
func (i *Int) Devar() Expr   { return i }
func (f Float) Devar() Expr  { return f }
func (r Record) Devar() Expr { return r }
func (u Union) Devar() Expr  { return u }
func (v Var) Devar() Expr    { return drop(v.Index, v.ctx).Arg.Devar() }
func (a *Appl) Devar() Expr  { return a }
func (a *Abst) Devar() Expr  { return a }
func (s Switch) Devar() Expr { return s }
func (r Ref) Devar() Expr    { return r }
func (g GoFunc) Devar() Expr { return g }

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
	return Union{u.Alternative, fields}
}
func (v Var) WithCtx(ctx *ctx) Expr { return Var{ctx, v.Index} }
func (a *Appl) WithCtx(ctx *ctx) Expr {
	return &Appl{
		reduced: a.reduced,
		ctx:     ctx,
		LDrop:   a.LDrop,
		RDrop:   a.RDrop,
		Left:    a.Left,
		Right:   a.Right,
	}
}
func (a *Abst) WithCtx(ctx *ctx) Expr  { return &Abst{ctx, a.Body} }
func (s Switch) WithCtx(ctx *ctx) Expr { return Switch{ctx, s.Expr, s.Cases} }
func (r Ref) WithCtx(ctx *ctx) Expr    { return r }
func (g GoFunc) WithCtx(*ctx) Expr     { return g }

func (c Char) Apply(Expr) Expr       { panic("not applicable") }
func (i *Int) Apply(Expr) Expr       { panic("not applicable") }
func (f Float) Apply(Expr) Expr      { panic("not applicable") }
func (r Record) Apply(Expr) Expr     { panic("not applicable") }
func (u Union) Apply(Expr) Expr      { panic("not applicable") }
func (v Var) Apply(Expr) Expr        { panic("not applicable") }
func (a *Appl) Apply(Expr) Expr      { panic("not applicable") }
func (a *Abst) Apply(arg Expr) Expr  { return a.Body.WithCtx(cons(arg, a.ctx)) }
func (s Switch) Apply(arg Expr) Expr { panic("not applicable") }
func (r Ref) Apply(arg Expr) Expr    { return (*r.Expr).Apply(arg) }
func (g GoFunc) Apply(arg Expr) Expr { return g(arg) }
