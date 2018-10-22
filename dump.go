package funky

import (
	"fmt"
	"io"

	"github.com/faiface/crux/runtime"
)

func dumpCodes(w io.Writer, globalIndices map[string][]int32, code *runtime.Code) {
	indicesToGlobals := make(map[int32]struct {
		Name  string
		Index int32
	})
	for name := range globalIndices {
		for i := range globalIndices[name] {
			indicesToGlobals[globalIndices[name][i]] = struct {
				Name  string
				Index int32
			}{name, int32(i)}
		}
	}

	var dumpIndented func(int, *runtime.Code)
	dumpIndented = func(indentation int, code *runtime.Code) {
		for i := 0; i < indentation; i++ {
			fmt.Fprintf(w, "  ")
		}

		switch code.Kind {
		case runtime.CodeValue:
			fmt.Fprintf(w, "VALUE %v\n", code.Value)
		case runtime.CodeOperator:
			fmt.Fprintf(w, "OPERATOR %s\n", runtime.OperatorString[code.X])
		case runtime.CodeMake:
			fmt.Fprintf(w, "MAKE %d\n", code.X)
		case runtime.CodeField:
			fmt.Fprintf(w, "FIELD %d\n", code.X)
		case runtime.CodeVar:
			fmt.Fprintf(w, "VAR %d\n", code.X)
		case runtime.CodeGlobal:
			global := indicesToGlobals[code.X]
			fmt.Fprintf(w, "GLOBAL %s/%d\n", global.Name, global.Index)
		case runtime.CodeAbst:
			fmt.Fprintf(w, "ABST %d\n", code.X)
		case runtime.CodeFastAbst:
			fmt.Fprintf(w, "FASTABST %d\n", code.X)
		case runtime.CodeAppl:
			fmt.Fprintf(w, "APPL\n")
		case runtime.CodeStrict:
			fmt.Fprintf(w, "STRICT\n")
		case runtime.CodeSwitch:
			fmt.Fprintf(w, "SWITCH\n")
		default:
			panic("invalid code")
		}

		for i := range code.Table {
			dumpIndented(indentation+1, &code.Table[i])
		}
	}

	dumpIndented(1, code)
}
