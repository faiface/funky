package runtime

type Thunk struct {
	Code *Code
	Data *Data
}

type CodeKind int16

const (
	CodeVar CodeKind = iota
	CodeAppl
	CodeStrictAppl
	CodeSwitch
	CodeGlobal
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

// reduceThunk efficiently evalutes the thunk decomposed in the arguments.
// Here are all possible kinds of results:
// 1. CodeVar, which does not evaluate to a thunk
// 3. CodeAbst
// 4. CodeGoFunc
// 5. CodeValue
func reduceThunk(code *Code, data *Data) (*Code, *Data) {
	for {
		switch code.Kind {
		case CodeVar:
			dropped := Drop(code.Drop, data)
			switch state := dropped.State.(type) {
			case *Thunk:
				vcode, vdata := reduceThunk(state.Code, state.Data)
				dropped.State = extractState(vcode, vdata)
				return vcode, vdata
			default:
				return code, data
			}

		case CodeAppl, CodeStrictAppl:
			dropped := Drop(code.Drop, data)
			lcode, ldata := reduceThunk(code.A, dropped)
			switch lcode.Kind {
			case CodeAbst:
				var arg State
				if code.Kind == CodeStrictAppl {
					rcode, rdata := reduceThunk(code.B, dropped)
					arg = extractState(rcode, rdata)
				} else {
					arg = &Thunk{code.B, dropped}
				}
				code, data = lcode.A, Cons(arg, Drop(lcode.Drop, ldata))
				continue
			case CodeGoFunc:
				return code, data
			default:
				panic("not an abstraction")
			}

		case CodeSwitch:
			dropped := Drop(code.Drop, data)
			ecode, edata := reduceThunk(code.A, dropped)
			union := extractState(ecode, edata).(*Union)
			bcode, bdata := &code.SwitchTable[union.Alternative], dropped
			for _, field := range union.Fields {
				bcode, bdata = reduceThunk(bcode, bdata)
				if bcode.Kind != CodeAbst {
					panic("switch case body not an abstraction")
				}
				bcode, bdata = bcode.A, Cons(field, Drop(bcode.Drop, bdata))
			}
			code, data = bcode, bdata
			continue

		case CodeGlobal:
			code = code.A
			data = nil

		case CodeAbst, CodeGoFunc:
			return code, data

		case CodeState:
			switch state := code.State.(type) {
			case *Thunk:
				vcode, vdata := reduceThunk(state.Code, state.Data)
				code.State = extractState(vcode, vdata)
				return vcode, vdata
			default:
				return code, data
			}
		}
	}
}

// extractState accepts a reduced thunk and returns its result
func extractState(code *Code, data *Data) State {
	switch code.Kind {
	case CodeVar:
		return Drop(code.Drop, data).State
	case CodeAbst:
		return &Thunk{code, data}
	case CodeGoFunc:
		return code.State.(GoFunc)(data)
	case CodeState:
		return code.State
	default:
		panic("not a reduced thunk")
	}
}

func (t *Thunk) Reduce() State {
	t.Code, t.Data = reduceThunk(t.Code, t.Data)
	return extractState(t.Code, t.Data)
}
