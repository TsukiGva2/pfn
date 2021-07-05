package pfn

import (
	"errors"
	"fmt"
	"strings"
)

const ( // Atypes
	/*
		types used for defining function arguments
		tNormal is for named arguments
		ex:

		.f(|x,y|->(+ x y))

		in this example both x and y
		will be of type tNormal, and defined as

		x, y = args[0], args[1]

		on the function body.

		tCompare is for pattern matched arguments
		ex:

		.f(|0|->"undefined")

		in this example there are no names to the
		arguments to be bind to, so the arguments
		are compared against the specified literal instead,
		you would see this on the function body:

		if 0 != args[0]:
			raise UnmatchedError()

		where '0' is of type tCompare.
	*/
	tNormal = iota
	tCompare
)

type argument struct {
	expr  string
	atype int
}

var ident int = 0
var fns = make(map[string]([]string))

type parserFn func() (string, error)

type Transpiler struct {
	current uint
	tokens  []Token
	Output  string
}

func (tp *Transpiler) start() {
	/*
		Initialize stuff and run
	*/

	ident = 0
	fns = make(map[string]([]string))

	tp.Output += "# this code was auto generated by pfn\n\n"

	tp.Output += "class UnmatchedError(Exception):\n\tpass\n\n"
	tp.Output += "class ArgcountError(Exception):\n\tpass\n\n"

	pfncall := "def __pfn_call(p, args):\n"
	pfncall += "	result=None\n"
	pfncall += "	broke=False\n"
	pfncall += "	for f in p:\n"
	pfncall += "		try:\n"
	pfncall += "			result=f(*args)\n"
	pfncall += "		except (UnmatchedError, ArgcountError):\n"
	pfncall += "			continue\n"
	pfncall += "		broke=True\n\n"
	pfncall += "		break\n"
	pfncall += "	if not broke:\n"
	pfncall += "		raise Exception('no matching function')\n\n"
	pfncall += "	return result\n"

	tp.Output += pfncall
	//tp.Output = fmt.Sprintf("broke=False\nfor f in [%v]:\n\ttry:\n\t\tf(%s)\n\texcept (UnmatchedError, ArgcountError):\n\t\tcontinue\n\tbroke=True\n\tbreak\n\nif not broke:\n\traise Exception('no matching function')\n")

	tp.run()
}

func (tp *Transpiler) run() {
	tp.Output += tp.code(cEOF, "")
}

func (tp *Transpiler) code(end int, alt string, extra ...parserFn) string {
	// (fn | var | expr)*

	var pfns []parserFn
	var output string

	pfns = append(pfns, extra...)
	pfns = append(pfns, tp.py, tp.out, tp.fn, tp.loop, tp.when, tp.variable, tp.expr, tp.class)

	for {
		tok := tp.ctoken()

		if tok.tokTy == end || tok.tokTy == cEOF || (tok.tokTy == cIdentifier && strings.Contains(alt, tok.lexeme)) {
			break
		}

		res, err := tp.rfwo(pfns)

		if err != nil {
			panic(err)
		}

		lines := strings.Split(res, "\n")

		for i := range lines {
			if !strings.HasPrefix(lines[i], "\t") {
				output += strings.Repeat("\t", ident) + lines[i] + "\n"
			} else {
				output += lines[i] + "\n"
			}
		}

		tp.advance(1)
	}

	return output
}

// parsing functions

func (tp *Transpiler) fn() (string, error) {
	// "." id "(" "|" (id ("," id)*)? "|" code ")"

	var fname string
	var code string
	var prefix string
	var args []argument

	tok := tp.ctoken()

	if tok.tokTy == cAt {
		prefix = "async "
		tp.advance(1)
	}

	tok = tp.ctoken()

	if tok.tokTy != cDot {
		return "", errors.New("not a function: no dot")
	}

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cIdentifier {
		return "", errors.New("not a function: no identifier")
	}

	fname = tok.lexeme

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cLparen {
		return "", errors.New("not a function: missing opening paren on arguments declaration")
	}

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cBOr {
		return "", errors.New("not a function: missing argument list")
	}

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cBOr {
		for {
			if tok.tokTy != cIdentifier {
				expr, err := tp.expr()

				if err != nil {
					return "", errors.New("not a function: error parsing expr")
				}

				args = append(args, argument{expr, tCompare})
			} else {
				args = append(args, argument{tok.lexeme, tNormal})
			}

			tp.advance(1)

			if tp.ctoken().tokTy != cComma {
				break
			}

			tp.advance(1)
			tok = tp.ctoken()
		}
	}

	tok = tp.ctoken()

	if tok.tokTy != cBOr {
		return "", errors.New("not a function: argument list not properly closed")
	}

	f, exists := fns[fname]
	oldName := fname

	if exists {
		fname = fmt.Sprintf("%s_%d", fname, len(f))
	}

	code = fmt.Sprintf("%sdef pfn_%s(*args):\n", prefix, fname)

	code += fmt.Sprintf("\tif len(args) < %d:\n\t\traise ArgcountError('too few arguments for function %s')\n", len(args), fname)

	for i := range args {
		arg := args[i]

		if arg.atype == tNormal {
			code += fmt.Sprintf("\t%s = args[%d]\n", arg.expr, i)
			continue
		}

		code += fmt.Sprintf("\tif %s != args[%d]:\n\t\traise UnmatchedError('unmatched')\n", arg.expr, i)
	}

	tp.advance(1)

	ident++
	code += tp.code(cRparen, "", tp.ret)
	ident--

	if exists {
		fns[oldName] = append(f, fmt.Sprintf("pfn_%s", fname))
		return code, nil
	}

	fns[oldName] = []string{fmt.Sprintf("pfn_%s", fname)}

	return code, nil
}

func (tp *Transpiler) ret() (string, error) {
	// "->" expr

	var output string

	tok := tp.ctoken()

	if tok.tokTy != cArrow {
		return "", errors.New("not a return statement: no arrow")
	}

	tp.advance(1)
	tok = tp.ctoken()

	output += "return "
	expr, err := tp.expr()

	if err != nil {
		return "", errors.New("not a return statement: error parsing expression")
	}

	output += expr

	return output, nil
}

func (tp *Transpiler) variable() (string, error) {
	// id ":=" expr

	var output string
	old := tp.current
	tok := tp.ctoken()

	ind, err := tp.index()

	if err != nil {
		tp.current = old
		if tok.tokTy != cIdentifier {
			return "", errors.New("not a variable: no identifier/index expr")
		}
		ind = tok.lexeme
	}

	output += ind

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cAssignment {
		return "", errors.New("not a variable: no assignment happening at all")
	}

	tp.advance(1)

	expr, err := tp.expr()

	if err != nil {
		return "", errors.New("not a variable: error parsing expr")
	}

	output += "=" + expr

	return output, nil
}

func (tp *Transpiler) expr() (string, error) {
	// (call | literal)

	fns := []parserFn{tp.let, tp.py, tp.call, tp.index, tp.ewhen, tp.list, tp.literal}
	old := tp.current

	for i := range fns {
		res, err := fns[i]()

		if err == nil {
			return res, nil
		}

		tp.current = old
	}

	return "", errors.New("error parsing expr")
}

func (tp *Transpiler) list() (string, error) {
	// "'"? "<" (expr ("," expr)*)? ">"

	var t []string = []string{"[", "]"}
	var output string

	tok := tp.ctoken()

	if tok.tokTy != cLt && tok.tokTy != cQuot {
		return "", errors.New("not a list: no opening < or '<")
	}

	if tok.tokTy == cQuot {
		tp.advance(1)
		tok = tp.ctoken()

		if tok.tokTy != cLt {
			return "", errors.New("not a tuple: no < after quote")
		}

		t = []string{"(", ")"}
	}

	tp.advance(1)
	tok = tp.ctoken()

	output += t[0]

	if tok.tokTy != cGt {
		for {
			expr, err := tp.expr()

			if err != nil {
				return "", errors.New("not a list: error parsing items")
			}

			output += expr

			tp.advance(1)
			tok = tp.ctoken()

			if tok.tokTy != cComma {
				break
			}

			output += ","

			tp.advance(1)
		}
	}

	tok = tp.ctoken()

	if tok.tokTy != cGt {
		return "", errors.New("not a list: no closing >")
	}

	tp.advance(1)
	tok = tp.ctoken()

	output += t[1]

	if tok.tokTy == cEloop && tok.lexeme == "where" {
		tp.advance(1)
		tok = tp.ctoken()

		if tok.tokTy != cIdentifier {
			return "", errors.New("not a list: error in 'where' clause, no identifier, did you mean 'where _'?")
		}

		v := tok.lexeme

		tp.advance(1)
		tok = tp.ctoken()

		if tok.tokTy == cEqArrow {
			tp.advance(1)
			tok = tp.ctoken()

			expr, err := tp.expr()

			if err != nil {
				return "", errors.New("not a list: (in 'where' clause) error parsing expr")
			}

			output = "[" + output + " for " + v + " in " + expr + "]"
			return output, nil
		}

		if tok.tokTy != cAssignment {
			return "", errors.New("not a list: error in 'where' clause, no ':=' or '=>'")
		}

		tp.advance(1)
		tok = tp.ctoken()

		expr, err := tp.expr()

		if err != nil {
			return "", errors.New("not a list: (in 'where' clause) error parsing expr")
		}

		output = "[" + output + " for " + v + " in " + "[" + expr + "]" + "][0]"
		tp.advance(1)
	}

	tp.previous(1)

	return output, nil
}

func (tp *Transpiler) literal() (string, error) {
	// (number | string | identifier)

	tok := tp.ctoken()

	switch tok.tokTy {
	case cNumber:
		break
	case cString:
		return "\"" + tok.literal.(string) + "\"", nil
	case cIdentifier:
		f, exists := fns[tok.lexeme]

		if exists {
			tok.lexeme = fmt.Sprintf("(lambda *args: __pfn_call([%s], args))", strings.Join(f, ","))
		}
	case cMinus:
		tp.advance(1)
		lit, err := tp.literal()
		if err != nil {
			return "", errors.New("not a valid literal of any kind")
		}
		return "-" + lit, nil

	default:
		return "", errors.New("not a valid literal of any kind")
	}

	return tok.lexeme, nil
}

func (tp *Transpiler) call() (string, error) {
	// "(" id (expr ("," expr)*)? ")"

	var isOp bool
	var unary bool
	var prefix string
	var output string

	tok := tp.ctoken()

	if tok.tokTy == cT {
		prefix = "await "
		tp.advance(1)
	}

	tok = tp.ctoken()

	if tok.tokTy != cLparen {
		return "", errors.New("not a function call: no opening paren")
	}

	tp.advance(1)
	tok = tp.ctoken()

	switch tok.tokTy {
	case cPlus:
		fallthrough
	case cMinus:
		fallthrough
	case cStar:
		fallthrough
	case cDoubleEq:
		fallthrough
	case cBangEq:
		fallthrough
	case cGt:
		fallthrough
	case cLAnd:
		fallthrough
	case cLOr:
		fallthrough
	case cLt:
		fallthrough
	case cGtEq:
		fallthrough
	case cLtEq:
		fallthrough
	case cSlash:
		isOp = true

	case cIdentifier:
		switch tok.lexeme {
		case "not":
			unary = true
		default:
			isOp = false
		}

	default:
		return "", errors.New("not a function call: no operator")
	}

	if unary {
		op := tok.lexeme
		tp.advance(1)
		tok = tp.ctoken()

		expr, err := tp.expr()

		if err != nil {
			return "", errors.New("not a function call: error parsing expr")
		}

		tp.advance(1)
		tok = tp.ctoken()

		if tok.tokTy != cRparen {
			return "", errors.New("not a function call: no closing paren")
		}

		output = "(" + op + " " + expr + ")"

		return output, nil
	}

	if isOp {
		output += "("

		op := tok.lexeme
		tp.advance(1)
		tok = tp.ctoken()

		/*if tok.tokTy != cLparen {
			return "", errors.New("not a function call: missing opening paren")
		}*/

		//tp.advance(1)
		//tok = tp.ctoken()

		if tok.tokTy != cRparen {

			argcount := 0

			for {
				expr, err := tp.expr()

				if err != nil {
					return "", errors.New("not a function call: error parsing arguments")
				}

				output += expr
				argcount++

				tp.advance(1)
				tok = tp.ctoken()

				if tok.tokTy == cEOF {
					return "", errors.New("general error (in tp.call): unclosed (")
				}

				if tok.tokTy == cRparen {
					break
				}

				output += " " + op + " "
			}

			if argcount < 1 {
				return "", errors.New("not a function call: not enough arguments to function " + op)
			}

			tok = tp.ctoken()

			output += ")"

			return output, nil
		}

		return "", errors.New("not a function call: not enough arguments to function " + op)
	}

	fname := tok.lexeme
	var args string

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cRparen {
		for {
			expr, err := tp.expr()

			if err != nil {
				return "", errors.New("not a function call: error parsing arguments")
			}

			args += expr

			tp.advance(1)
			tok = tp.ctoken()

			if tok.tokTy == cEOF {
				return "", errors.New("general error (in tp.call): unclosed (")
			}

			if tok.tokTy == cRparen {
				break
			}

			args += ","
		}
	}

	tok = tp.ctoken()

	if tok.tokTy != cRparen {
		return "", errors.New("not a function call: missing closing )")
	}

	f, exists := fns[fname]

	if exists {
		output = fmt.Sprintf(prefix+"__pfn_call([%s], [%s])", strings.Join(f, ","), args)
		//output = fmt.Sprintf("broke=False\nfor f in [%v]:\n\ttry:\n\t\tf(%s)\n\texcept (UnmatchedError, ArgcountError):\n\t\tcontinue\n\tbroke=True\n\tbreak\n\nif not broke:\n\traise Exception('no matching function')\n", strings.Join(f, ","), args)

		return output, nil
	}

	output = fmt.Sprintf("%s(%s)", fname, args)

	return output, nil
}

func (tp *Transpiler) ewhen() (string, error) {
	// "when" expr "do" code "end"
	var output string
	tok := tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "use" {
		return "", errors.New("not a when expression: missing 'use'")
	}

	tp.advance(1)

	expr, err := tp.expr()

	if err != nil {
		return "", errors.New("not a when expr: error parsing expression")
	}

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "when" {
		return "", errors.New("not a when expr: no 'when'")
	}

	tp.advance(1)

	output += expr + " if "

	expr, err = tp.expr()

	if err != nil {
		return "", errors.New("not a when expr: error parsing expression")
	}

	output += expr

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "else" {
		tp.previous(1)
		return output, nil
	}

	tp.advance(1)
	tok = tp.ctoken()

	expr, err = tp.expr()

	if err != nil {
		return "", errors.New("not a when expr: error parsing expr")
	}

	output += " else " + expr

	return output, nil
}

func (tp *Transpiler) when() (string, error) {
	// "when" expr "do" code "end"
	var output string
	tok := tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "when" {
		return "", errors.New("not a when block: missing 'when'")
	}

	output += "if "

	tp.advance(1)

	expr, err := tp.expr()

	if err != nil {
		return "", errors.New("not a when block: error parsing expression")
	}

	output += expr

	tp.advance(1)
	tok = tp.ctoken()

	output += ":\n"

	if tok.tokTy != cIdentifier || tok.lexeme != "do" {
		return "", errors.New("not a when block: no 'do'")
	}

	tp.advance(1)

	ident++
	code := tp.code(cEnd, "else", tp.ret)
	ident--

	output += code

	tok = tp.ctoken()

	if tok.tokTy == cIdentifier && tok.lexeme == "else" {
		output += "else"

		tp.advance(1)
		tok = tp.ctoken()

		output += ":\n"

		ident++
		code = tp.code(cEnd, "", tp.ret)
		ident--

		output += code
	}

	return output, nil
}

func (tp *Transpiler) py() (string, error) {
	// "py" "{" any "}"
	var output string

	tok := tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "py" {
		return "", errors.New("not a python block: missing 'py' keyword")
	}

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cString {
		return "", errors.New("not a python block: missing string")
	}

	output += tok.literal.(string)

	return output, nil
}

func (tp *Transpiler) loop() (string, error) {
	// "loop" code "where" id ("," id)* ("=>"|":=") expr
	var output string
	var lvar string

	tok := tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "loop" {
		return "", errors.New("not a loop: no 'loop' keyword")
	}

	tp.advance(1)

	ident++
	code := tp.code(cEloop, "", tp.brk)
	ident--

	tok = tp.ctoken()

	if tok.tokTy != cEloop || tok.lexeme != "where" {
		if tok.tokTy == cEloop && tok.lexeme == "while" {
			tp.advance(1)

			expr, err := tp.expr()

			if err != nil {
				return "", errors.New("not a loop: error parsing expr")
			}

			output = fmt.Sprintf("while %s:\n%s", expr, code)
			return output, nil
		}

		return "", errors.New("not a loop: no where or when")
	}

	tp.advance(1)
	tok = tp.ctoken()

	for {
		if tok.tokTy != cIdentifier {
			return "", errors.New("not a loop: no identifier after where or comma")
		}

		lvar += tok.lexeme

		tp.advance(1)
		tok = tp.ctoken()

		if tok.tokTy != cComma {
			break
		}

		lvar += ","

		tp.advance(1)
		tok = tp.ctoken()
	}

	if tok.tokTy == cAssignment {
		tp.advance(1)

		expr, err := tp.expr()

		if err != nil {
			return "", errors.New("not a loop: error parsing expr")
		}

		output = fmt.Sprintf("%s=%s\nwhile %s:\n%s", lvar, expr, lvar, code)
		return output, nil
	}

	if tok.tokTy != cEqArrow {
		return "", errors.New("not a loop: no '=>' or ':='")
	}

	tp.advance(1)
	tok = tp.ctoken()

	expr, err := tp.expr()

	if err != nil {
		return "", errors.New("not a loop: error parsing expr")
	}

	output = fmt.Sprintf("for %s in %s:\n%s", lvar, expr, code)

	return output, nil
}

func (tp *Transpiler) brk() (string, error) {
	tok := tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "break" {
		return "", errors.New("not break")
	}

	return "break", nil
}

func (tp *Transpiler) let() (string, error) {
	tok := tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "let" {
		return "", errors.New("not a let clause")
	}

	tp.advance(1)

	tok = tp.ctoken()

	v, err := tp.variable()

	if err != nil {
		return "", err
	}

	tp.advance(1)
	varname := tok.lexeme

	code := v
	code += "\n"

	tok = tp.ctoken()
	if tok.tokTy != cIdentifier || tok.lexeme != "in" {
		return "", errors.New("not a let clause: no in")
	}

	tp.advance(1)

	code += tp.code(cEnd, "", tp.ret)

	code += "del " + varname
	code += "\n"
	return code, nil
}

func (tp *Transpiler) index() (string, error) {
	// (id|list) ":" expr
	var output string = ""
	var literal string
	tok := tp.ctoken()

	if tok.tokTy != cIdentifier {
		list, err := tp.list()

		if err != nil {
			return "", errors.New("not an index expr: no list or identifier")
		}

		literal = list
	} else {
		literal = tok.lexeme
	}

	output += literal

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cColon {
		return "", errors.New("not an index expr: no colon")
	}

	tp.advance(1)
	tok = tp.ctoken()

	expr, err := tp.expr()
	if err != nil {
		return "", errors.New("not an index expr: error parsing expr")
	}

	output += "[" + expr + "]"

	return output, nil
}

func (tp *Transpiler) class() (string, error) {
	// "=" id "(" code ")"
	var name string
	var fields []string

	tok := tp.ctoken()

	if tok.tokTy != cEq {
		return "", errors.New("not a class: no =")
	}

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cIdentifier {
		return "", errors.New("not a class: no identifier")
	}
	name = tok.lexeme

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cLparen {
		return "", errors.New("not a class: no '{'")
	}

	tp.advance(1)
	tok = tp.ctoken()

	for true {
		if tok.tokTy == cRparen || tok.tokTy == cEOF {
			break
		}

		if tok.tokTy != cIdentifier {
			return "", errors.New("not a class: no identifier")
		}

		fields = append(fields, tok.lexeme)

		tp.advance(1)
		tok = tp.ctoken()
	}

	if tok.tokTy != cRparen {
		return "", errors.New("not a class: unexpected EOF")
	}

	list := ""

	for i := range fields {
		list += fields[i]

		if i < len(fields)-1 {
			list += ","
		}
	}

	output := "class "+name+":\n"
	ident++
	output += strings.Repeat("\t", ident)
	output += "def __init__(self,"+list+"):\n"
	ident++

	for i := range fields {
		output += strings.Repeat("\t", ident)
		output += "self." + fields[i] + "=" + fields[i] + "\n"
	}

	ident -= 2

	output += "\n"

	return output, nil
}

// dangerous language constructs

func (tp *Transpiler) out() (string, error) {
	tok := tp.ctoken()

	if tok.tokTy != cIdentifier || tok.lexeme != "out!" {
		return "", errors.New("not an out statement: no 'out!'")
	}

	tp.advance(1)

	code := tp.code(cEnd, "")

	tp.Output = code + "\n" + tp.Output

	return "", nil
}

// utils

func (tp *Transpiler) advance(n uint) {
	if tp.current+n < uint(len(tp.tokens)) {
		tp.current += n
	}
}

func (tp *Transpiler) previous(n uint) {
	if tp.current >= n {
		tp.current -= n
	}
}

func (tp Transpiler) ctoken() Token {
	return tp.tokens[tp.current]

}
func (tp Transpiler) err(where string, msg string) {
	//if !haderror {
	//haderror = true
	fmt.Printf("had error on %s, line %d, col %d\n%s\n",
		where, tp.ctoken().line+1, (tp.ctoken().col+1)/(tp.ctoken().line+1), msg)
	//}
}

func (tp *Transpiler) rfwo(fns []parserFn) (string, error) {
	/*
		Return First Working One

		returns the first matching
		parsing function on the given
		list
	*/

	old := tp.current

	var last error

	for i := range fns {
		res, err := fns[i]()

		if err == nil {
			return res, nil
		}

		last = err
		tp.current = old
	}

	tp.err("rfwo", "no matching pfn")
	panic(last)
}
