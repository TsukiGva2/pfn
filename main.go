package pfn

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

//var haderror bool = false

var outfile string
var libdir string

func main() {
	var filename string

	flag.StringVar(&filename, "c", "", "file name to compile")
	flag.StringVar(&outfile, "o", "out.py", "out file name")
	flag.StringVar(&libdir, "l", "lib/", "path to libraries dir")

	flag.Parse()

	if filename == "" {
		repl()
		return
	}

	runFile(filename, true)
}

func runFile(f string, w bool) string {
	cont, err := ioutil.ReadFile(f)

	if err != nil {
		panic(err)
	}

	code := string(cont)
	t := Run(code, false)

	if w {
		err = ioutil.WriteFile(outfile, []byte(t.Output), 0654)
	}

	if err != nil {
		log.Fatal(err)
	}

	return t.Output
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
			Run(input, true)
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
				Run(line, true)
				input = ""
				//haderror = false
			} else {
				input += line + "\n"
			}
		}
	}
}

func Run(code string, p bool) Transpiler {
	cont, err := ioutil.ReadFile(libdir + "prelude.pfn")

	if err != nil {
		panic(err)
	}

	code = string(cont) + "\n\n" + code


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
