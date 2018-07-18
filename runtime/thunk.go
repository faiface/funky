package runtime

type Thunk struct {
	Code *Code
	Data *Data
	Memo Value
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
	CodeValue
)

type Code struct {
	Kind        CodeKind
	Drop        int32
	A, B        *Code
	SwitchTable []Code
	Value       Value
}

type Data struct {
	Value Value
	Next  *Data
}

func Cons(x Value, d *Data) *Data {
	return &Data{x, d}
}

func Drop(n int32, d *Data) *Data {
	for i := int32(0); i < n; i++ {
		d = d.Next
	}
	return d
}

var Reductions = 0

// reduceThunk efficiently evalutes the thunk decomposed in the arguments
func reduceThunk(code *Code, data *Data) (*Code, *Data, Value) {
	for {
		Reductions++

		switch code.Kind {
		case CodeVar:
			value := Drop(code.Drop, data).Value.Reduce()
			switch value := value.(type) {
			case *Thunk:
				return value.Code, value.Data, value
			default:
				return code, data, value
			}

		case CodeAppl, CodeStrictAppl:
			dropped := Drop(code.Drop, data)
			lcode, ldata, left := reduceThunk(code.A, dropped)
			switch lcode.Kind {
			case CodeAbst:
				var arg Value
				if code.Kind == CodeStrictAppl {
					var (
						acode *Code
						adata *Data
					)
					acode, adata, arg = reduceThunk(code.B, dropped)
					if acode.Kind == CodeAbst {
						arg = &Thunk{acode, adata, nil}
					}
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
			return code, data, code.Value.(GoFunc)(data)

		case CodeValue:
			return code, data, code.Value.Reduce()
		}
	}
}

func (t *Thunk) Reduce() Value {
	if t.Memo != nil {
		return t.Memo
	}
	t.Code, t.Data, t.Memo = reduceThunk(t.Code, t.Data)
	if t.Code.Kind == CodeAbst {
		return t
	}
	t.Code, t.Data = nil, nil
	return t.Memo
}
