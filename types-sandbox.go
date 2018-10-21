package funky

import (
	"bufio"
	"fmt"
	"os"

	"github.com/faiface/funky/parse"

	"github.com/faiface/funky/compile"
)

func runTypesSandbox(env *compile.Env) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		code := scanner.Text()
		tokens, err := parse.Tokenize("sandbox", code)
		if err != nil {
			fmt.Println(err)
			continue
		}
		exp, err := parse.Expr(tokens)
		if err != nil {
			fmt.Println(err)
			continue
		}
		results, err := env.TypeInferExpr(exp)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, result := range results {
			fmt.Println(result.Type)
		}
	}
}
