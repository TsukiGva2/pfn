package pfn

import (
	"fmt"
	"io/ioutil"
)

var utils string = `
.Class (|name|
	(exec (+ "class " name ":\n" "\tdef define(m):\n\t\texec('" name ".' + m + ' = ' + 'pfn_' + m + 'at" name "', globals())\n") (globals))
)
`

func Run(code string, p bool, noprelude bool) Transpiler {
	code = utils + code
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
