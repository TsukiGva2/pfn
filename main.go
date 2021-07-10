package pfn

import (
	"fmt"
	"io/ioutil"
)

func Run(code string, p bool, noprelude bool, customErrs ...string) Transpiler {
	cerrs = customErrs
	if !noprelude {
		cont, err := ioutil.ReadFile("prelude.pfn")

		if err != nil {
			panic(err)
		}

		code = string(cont) + "\n\n" + code
	}

	sc := Scanner{code, 0, 0, 0, false}
	tokens := sc.scanTokens()
	//for i := range tokens {
	//  fmt.Printf("%#v\n", tokens[i])
	//}
	tp := Transpiler{0, tokens, ""}
	tp.start()

	if p {
		fmt.Print(tp.Output)
	}

	return tp
}
