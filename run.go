package funky

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/faiface/funky/compile"
	"github.com/faiface/funky/parse"
	"github.com/faiface/funky/runtime"

	cxr "github.com/faiface/crux/runtime"
)

func Run(main string) (value *runtime.Value, cleanup func()) {
	stats := flag.Bool("stats", false, "print stats after running program")
	typesSandbox := flag.Bool("types", false, "start types sandbox instead of running the program")
	dump := flag.String("dump", "", "specify a file to dump the compiled code into")
	flag.Parse()

	compilationStart := time.Now()

	var definitions []parse.Definition

	for _, filename := range flag.Args() {
		b, err := ioutil.ReadFile(filename)
		handleErrs(err)
		tokens, err := parse.Tokenize(filename, string(b))
		handleErrs(err)
		defs, err := parse.Definitions(tokens)
		handleErrs(err)
		definitions = append(definitions, defs...)
	}

	env := new(compile.Env)
	for _, definition := range definitions {
		err := env.Add(definition)
		handleErrs(err)
	}

	errs := env.Validate()
	handleErrs(errs...)

	if *typesSandbox {
		runTypesSandbox(env)
		os.Exit(0)
	}

	errs = env.TypeInfer()
	handleErrs(errs...)
	globalIndices, globalValues, codeIndices, codes := env.Compile(main)

	if len(globalIndices[main]) == 0 {
		handleErrs(fmt.Errorf("no %s function", main))
	}
	if len(globalIndices[main]) > 1 {
		handleErrs(fmt.Errorf("multiple %s functions", main))
	}

	if *dump != "" {
		df, err := os.Create(*dump)
		handleErrs(err)
		for name := range globalIndices {
			for i := range globalIndices[name] {
				fmt.Fprintf(df, "# %v\n", env.SourceInfo(name, i))
				fmt.Fprintf(df, "# %v\n", env.TypeInfo(name, i))
				fmt.Fprintf(df, "FUNC %s/%d\n", name, i)
				dumpCodes(df, globalIndices, &codes[codeIndices[name][i]])
				fmt.Fprintln(df)
			}
		}
		handleErrs(df.Close())
	}

	program := &runtime.Value{Globals: globalValues, Value: globalValues[globalIndices[main][0]]}

	runningStart := time.Now()

	return program, func() {
		if *stats {
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "STATS\n")
			fmt.Fprintf(os.Stderr, "reductions:       %d\n", cxr.Reductions)
			fmt.Fprintf(os.Stderr, "compilation time: %v\n", runningStart.Sub(compilationStart))
			fmt.Fprintf(os.Stderr, "running time:     %v\n", time.Since(runningStart))
		}
	}
}

func handleErrs(errs ...error) {
	bad := false
	for _, err := range errs {
		if err != nil {
			bad = true
			fmt.Fprintln(os.Stderr, err)
		}
	}
	if bad {
		os.Exit(1)
	}
}
