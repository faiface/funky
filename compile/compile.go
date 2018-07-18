package compile

import (
	"fmt"

	"github.com/faiface/funky/expr"
	"github.com/faiface/funky/runtime"
	"github.com/faiface/funky/types"
	"github.com/faiface/funky/types/typecheck"
)

func (env *Env) Compile(main string) (*runtime.Box, error) {
	var (
		offsets = make(map[string][]int)

		codes []runtime.Code

		linkAs           []struct{ from, to int }
		linkBs           []struct{ from, to int }
		linkSwitchTables []struct{ from, to, len int }
		linkRefs         []struct {
			from  int
			to    string
			index int
		}
	)

	// gather function type info
	global := make(map[string][]types.Type)
	for name, impls := range env.funcs {
		for _, impl := range impls {
			global[name] = append(global[name], impl.TypeInfo())
		}
	}

	// here's a recursive helper function for compiling expressions
	// it's local here, for easy access to env fields and local variables in this function
	var compileExpr func([]string, expr.Expr) int
	compileExpr = func(locals []string, e expr.Expr) int {
		drop := int32(0)
		for _, local := range locals {
			if e.HasFree(local) {
				break
			}
			drop++
		}

		switch e := e.(type) {
		case *expr.Var:
			// local variable
			if int(drop) < len(locals) {
				offset := len(codes)
				codes = append(codes, runtime.Code{
					Kind: runtime.CodeVar,
					Drop: drop,
				})
				return offset
			}
			// global function reference
			for i, typ := range global[e.Name] {
				if typecheck.CheckIfUnify(env.names, e.TypeInfo(), typ) {
					offset := len(codes)
					codes = append(codes, runtime.Code{
						Kind: runtime.CodeRef,
						Drop: drop,
					})
					linkRefs = append(linkRefs, struct {
						from  int
						to    string
						index int
					}{offset, e.Name, i})
					return offset
				}
			}
			panic("unreachable")

		case *expr.Appl:
			offset := len(codes)
			codes = append(codes, runtime.Code{
				Kind: runtime.CodeAppl,
				Drop: drop,
			})
			left := compileExpr(locals[drop:], e.Left)
			right := compileExpr(locals[drop:], e.Right)
			linkAs = append(linkAs, struct{ from, to int }{offset, left})
			linkBs = append(linkBs, struct{ from, to int }{offset, right})
			return offset

		case *expr.Abst:
			offset := len(codes)
			codes = append(codes, runtime.Code{
				Kind: runtime.CodeAbst,
				Drop: drop,
			})
			newLocals := append([]string{e.Bound.Name}, locals[drop:]...)
			body := compileExpr(newLocals, e.Body)
			linkAs = append(linkAs, struct{ from, to int }{offset, body})
			return offset

		case *expr.Switch:
			offset := len(codes)
			codes = append(codes, runtime.Code{
				Kind: runtime.CodeSwitch,
				Drop: drop,
			})
			exp := compileExpr(locals[drop:], e.Expr)
			linkAs = append(linkAs, struct{ from, to int }{offset, exp})

			switchTableOffset := len(codes)
			codes = append(codes, make([]runtime.Code, len(e.Cases))...)
			for i := 0; i < len(e.Cases); i++ {
				codes[switchTableOffset+i] = runtime.Code{
					Kind: runtime.CodeRef,
					Drop: 0,
				}
			}
			linkSwitchTables = append(linkSwitchTables, struct {
				from, to, len int
			}{offset, switchTableOffset, len(e.Cases)})

			for i := range e.Cases {
				cas := compileExpr(locals[drop:], e.Cases[i].Body)
				linkAs = append(linkAs, struct{ from, to int }{switchTableOffset + i, cas})
			}

			return offset

		case *expr.Char:
			offset := len(codes)
			codes = append(codes, runtime.Code{
				Kind:  runtime.CodeValue,
				Value: &runtime.Char{Value: e.Value},
			})
			return offset

		case *expr.Int:
			offset := len(codes)
			codes = append(codes, runtime.Code{
				Kind:  runtime.CodeValue,
				Value: &runtime.Int{Value: e.Value},
			})
			return offset

		case *expr.Float:
			offset := len(codes)
			codes = append(codes, runtime.Code{
				Kind:  runtime.CodeValue,
				Value: &runtime.Float{Value: e.Value},
			})
			return offset
		}

		panic("unreachable")
	}

	// compile individual functions
	for name, impls := range env.funcs {
		for _, impl := range impls {
			switch impl := impl.(type) {
			case *internalFunc:
				offsets[name] = append(offsets[name], len(codes))
				for j := 0; j < impl.Arity; j++ {
					codes = append(codes, runtime.Code{
						Kind: runtime.CodeAbst,
					})
					linkAs = append(linkAs, struct{ from, to int }{len(codes) - 1, len(codes)})
				}
				codes = append(codes, runtime.Code{
					Kind:  runtime.CodeGoFunc,
					Value: impl.GoFunc,
				})

			case *exprFunc:
				//TODO: proper sharing of top-level functions
				offset := compileExpr(nil, impl.Expr)
				offsets[name] = append(offsets[name], offset)
			}
		}
	}

	// finalize all direct links to codes and switch tables
	for _, link := range linkAs {
		codes[link.from].A = &codes[link.to]
	}
	for _, link := range linkBs {
		codes[link.from].B = &codes[link.to]
	}
	for _, link := range linkSwitchTables {
		codes[link.from].SwitchTable = codes[link.to : link.to+link.len]
	}
	for _, link := range linkRefs {
		codes[link.from].A = &codes[offsets[link.to][link.index]]
	}

	if len(env.funcs[main]) == 0 {
		return nil, &Error{
			SourceInfo: nil,
			Msg:        fmt.Sprintf("no %s function", main),
		}
	}
	if len(env.funcs[main]) > 1 {
		return nil, &Error{
			SourceInfo: nil,
			Msg:        fmt.Sprintf("multiple %s functions", main),
		}
	}

	return &runtime.Box{Value: &runtime.Thunk{
		Code: &codes[offsets[main][0]],
		Data: nil,
	}}, nil
}
