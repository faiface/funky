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
	errs = env.TypeInfer()
	handleErrs(errs...)
	program, err := env.Compile(main)
	handleErrs(err)

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
	if len(errs) == 0 {
		return
	}
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}
