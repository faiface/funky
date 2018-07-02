package runtime

type Thunk struct {
	Code *Code
	Data *Data
	Memo State
}

type CodeKind int16

const (
	CodeVar CodeKind = iota
	CodeAppl
	CodeStrictAppl
	CodeSwitch
	CodeRef
	CodeAbst
	CodeGoFunc
	CodeState
)

type Code struct {
	Kind        CodeKind
	Drop        int32
	A, B        *Code
	SwitchTable []Code
	State       State
}

type Data struct {
	State State
	Next  *Data
}

func Cons(x State, d *Data) *Data {
	return &Data{x, d}
}

func Drop(n int32, d *Data) *Data {
	for i := int32(0); i < n; i++ {
		d = d.Next
	}
	return d
}

var Reductions = 0

// reduceThunk efficiently evalutes the thunk decomposed in the arguments.
// Here are all possible kinds of results:
// 1. CodeVar, which does not evaluate to a thunk
// 3. CodeAbst
// 4. CodeGoFunc
// 5. CodeValue
func reduceThunk(code *Code, data *Data) (*Code, *Data, State) {
	for {
		Reductions++

		switch code.Kind {
		case CodeVar:
			state := Drop(code.Drop, data).State.Reduce()
			switch state := state.(type) {
			case *Thunk:
				return state.Code, state.Data, state
			default:
				return code, data, state
			}

		case CodeAppl, CodeStrictAppl:
			dropped := Drop(code.Drop, data)
			lcode, ldata, left := reduceThunk(code.A, dropped)
			switch lcode.Kind {
			case CodeAbst:
				var arg State
				if code.Kind == CodeStrictAppl {
					_, _, arg = reduceThunk(code.B, dropped)
				} else {
					arg = &Thunk{code.B, dropped, nil}
				}
				code, data = lcode.A, Cons(arg, Drop(lcode.Drop, ldata))
				continue
			case CodeGoFunc:
				code, data = lcode, Cons(&Thunk{code.B, dropped, nil}, Drop(lcode.Drop, ldata))
				return code, data, left.(GoFunc)(data)
			default:
				panic("not an abstraction")
			}

		case CodeSwitch:
			dropped := Drop(code.Drop, data)
			_, _, expr := reduceThunk(code.A, dropped)
			union := expr.(*Union)
			bcode, bdata := &code.SwitchTable[union.Alternative], dropped
			for _, field := range union.Fields {
				bcode, bdata, _ = reduceThunk(bcode, bdata)
				if bcode.Kind != CodeAbst {
					panic("switch case body not an abstraction")
				}
				bcode, bdata = bcode.A, Cons(field, Drop(bcode.Drop, bdata))
			}
			code, data = bcode, bdata
			continue

		case CodeRef:
			dropped := Drop(code.Drop, data)
			code = code.A
			data = dropped
			continue

		case CodeAbst:
			return code, data, nil

		case CodeGoFunc:
			return code, data, code.State.(GoFunc)(data)

		case CodeState:
			return code, data, code.State.Reduce()
		}
	}
}

func (t *Thunk) Reduce() State {
	if t.Memo != nil {
		return t.Memo
	}
	t.Code, t.Data, t.Memo = reduceThunk(t.Code, t.Data)
	if t.Code.Kind == CodeAbst {
		return t
	}
	return t.Memo
}
