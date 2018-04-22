package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/faiface/funky/compile"
	"github.com/faiface/funky/parse"
	"github.com/faiface/funky/runtime"
)

func main() {
	filename := "test.fn"
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	tokens, err := parse.Tokenize(filename, string(b))
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
	program, err := env.Compile("main")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	in, out := bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout)
loop:
	for {
		switch program.Alternative() {
		case 0: // done
			out.Flush()
			break loop
		case 1: // putc
			out.WriteRune(program.Field(0).Char())
			program = program.Field(1)
		case 2: // getc
			out.Flush()
			r, _, err := in.ReadRune()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			program = program.Field(0).Apply(runtime.MkChar(r))
		}
	}
}
