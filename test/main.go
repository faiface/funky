package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/faiface/funky/compile"
	"github.com/faiface/funky/parse"
)

func main() {
	filename := "test.fn"
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	tokens := parse.Tokenize(filename, string(b))
	defs, err := parse.Definitions(tokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	env := new(compile.Env)
	for _, def := range defs {
		err := env.Add(def)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
	errs := env.Validate()
	if errs != nil {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	errs = env.TypeInfer()
	if errs != nil {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
}
