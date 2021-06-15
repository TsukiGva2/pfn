package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

//var haderror bool = false

var outfile string = "out.py"

func main() {
	if len(os.Args) > 1 {
		if len(os.Args) > 2 {
			outfile = os.Args[2]
		}

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
	t := run(code, false)

	err = ioutil.WriteFile(outfile, []byte(t.output), 0654)

	if err != nil {
		log.Fatal(err)
	}

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
			run(input, true)
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
				run(line, true)
				input = ""
				//haderror = false
			} else {
				input += line + "\n"
			}
		}
	}
}

func run(code string, p bool) Transpiler {
	sc := Scanner{code, 0, 0, 0}
	tokens := sc.scanTokens()
	//for i := range tokens {
	//  fmt.Printf("%#v\n", tokens[i])
	//}
	tp := Transpiler{0, tokens, ""}
	tp.start()

	if p {
		fmt.Print(tp.output)
	}

	return tp
}
