package funky

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/faiface/funky/compile"
	"github.com/faiface/funky/parse"
	"github.com/faiface/funky/runtime"
)

func Run(main string) *runtime.Value {
	flag.Parse()

	var definitions []parse.Definition

	for _, filename := range flag.Args() {
		b, err := ioutil.ReadFile(filename)
		handleErr(err)
		tokens, err := parse.Tokenize(filename, string(b))
		handleErr(err)
		defs, err := parse.Definitions(tokens)
		handleErr(err)
		definitions = append(definitions, defs...)
	}

	env := new(compile.Env)
	for _, definition := range definitions {
		err := env.Add(definition)
		handleErr(err)
	}

	errs := env.Validate()
	handleErrs(errs)
	errs = env.TypeInfer()
	handleErrs(errs)
	program, err := env.Compile(main)
	handleErr(err)

	return program
}

func handleErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handleErrs(errs []error) {
	if len(errs) == 0 {
		return
	}
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}
