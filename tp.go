package main

import (
	"errors"
	"fmt"
	"strings"
)

const (
	tNormal = iota
	tCompare
)

var ident int = 0
var fns = make(map[string]([]string))

type parserFn func() (string, error)

type Transpiler struct {
	current uint
	tokens  []Token
	output  string
}

type argument struct {
	expr  string
	atype int
}

func (tp *Transpiler) start() {
	/*
		Initialize stuff and run
	*/

	ident = 0

	tp.output += "# this code was auto generated by pfn\n\n"

	tp.output += "class UnmatchedError(Exception):\n\tpass\n\n"
	tp.output += "class ArgcountError(Exception):\n\tpass\n\n"

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

	tp.output += pfncall
	//tp.output = fmt.Sprintf("broke=False\nfor f in [%v]:\n\ttry:\n\t\tf(%s)\n\texcept (UnmatchedError, ArgcountError):\n\t\tcontinue\n\tbroke=True\n\tbreak\n\nif not broke:\n\traise Exception('no matching function')\n")

	tp.run()
}

func (tp *Transpiler) run() {
	tp.output += tp.code(cEOF, "")
}

func (tp *Transpiler) code(end int, alt string, extra ...parserFn) string {
	// (fn | var | expr)*

	var pfns []parserFn
	var output string

	pfns = append(pfns, extra...)
	pfns = append(pfns, tp.py, tp.fn, tp.when, tp.variable, tp.expr)

	for {
		tok := tp.ctoken()

		if tok.tokTy == end || tok.tokTy == cEOF || (tok.tokTy == cIdentifier && tok.lexeme == alt) {
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
	var args []argument

	tok := tp.ctoken()

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

	code = fmt.Sprintf("def pfn_%s(*args):\n", fname)

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

	tp.advance(1)
	tok := tp.ctoken()

	if tok.tokTy != cAssignment {
		return "", errors.New("not a variable: no assignment happening at all")
	}

	tp.previous(1)
	tok = tp.ctoken()

	if tok.tokTy != cIdentifier {
		return "", errors.New("not a variable: no identifier")
	}

	output += tok.lexeme

	tp.advance(2)

	expr, err := tp.expr()

	if err != nil {
		return "", errors.New("not a variable: error parsing expr")
	}

	output += "=" + expr

	return output, nil
}

func (tp *Transpiler) expr() (string, error) {
	// (call | literal)

	fns := []parserFn{tp.py, tp.call, tp.ewhen, tp.list, tp.literal}
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

	if tok.tokTy == cIdentifier && tok.lexeme == "where" {
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
		break
	case cIdentifier:
		break

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
	var output string

	tok := tp.ctoken()

	if tok.tokTy != cLparen {
		return "", errors.New("not a function call: no opening paren")
	}

	tp.advance(1)
	tok = tp.ctoken()

	switch tok.tokTy {
	case cIdentifier:
		isOp = false

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
	case cLt:
		fallthrough
	case cGtEq:
		fallthrough
	case cLtEq:
		fallthrough
	case cSlash:
		isOp = true

	default:
		return "", errors.New("not a function call: no operator")
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

				output += op
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
		output = fmt.Sprintf("__pfn_call([%s], [%s])", strings.Join(f, ","), args)
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

	if tok.tokTy != cLbrace {
		return "", errors.New("not a python block: missing opening brace")
	}

	tp.advance(1)
	tok = tp.ctoken()

	for tok.tokTy != cRbrace {
		output += tok.lexeme + " "

		tp.advance(1)
		tok = tp.ctoken()

		if tok.tokTy == cEOF {
			return output, nil
		}
	}

	return output, nil
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
