package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

//var haderror bool = false

func main() {
	if len(os.Args) > 1 {
		runFile(os.Args[1])
		return
	}

	repl()
}

func runFile(f string) string {
	cont, err := ioutil.ReadFile(f)

	if err != nil {
		panic(err)
	}

	code := string(cont)
	t := run(code)

	return t.output
}

func repl() {
	var input string
	var automode bool
	input = ""

	for {
		var line string
		fmt.Print("> ")
		sc := bufio.NewScanner(os.Stdin)
		if sc.Scan() {
			line = sc.Text()
		}

		switch line {
		case "run":
			run(input)
			line = ""
			input = ""
			//haderror = false
		case "list":
			fmt.Print(input)
		case "auto":
			automode = !automode
			fmt.Printf("automode is %v\n", automode)
		case "exit":
			return
		default:
			if automode {
				run(line)
				input = ""
				//haderror = false
			} else {
				input += line + "\n"
			}
		}
	}
}

func run(code string) Transpiler {
	sc := Scanner{code, 0, 0, 0}
	tokens := sc.scanTokens()
	//for i := range tokens {
	//  fmt.Printf("%#v\n", tokens[i])
	//}
	tp := Transpiler{0, tokens, ""}
	tp.start()

	fmt.Println(tp.output)

	return tp
}
