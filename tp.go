package main

import (
	"errors"
	"fmt"
	"strings"
)

var ident int = 0

type parserFn func() (string, error)

type Transpiler struct {
	current uint
	tokens  []Token
	output  string
}

func (tp *Transpiler) start() {
	/*
		Initialize stuff and run
	*/

	tp.run()
}

func (tp *Transpiler) run() {
	tp.output += tp.code(cEOF)
}

func (tp *Transpiler) code(end int, extra ...parserFn) string {
	// (fn | var | expr)*

	var pfns []parserFn
	var output string

	pfns = append(pfns, extra...)
	pfns = append(pfns, tp.fn, tp.variable, tp.expr)

	for {
		tok := tp.ctoken()

		if tok.tokTy == end || tok.tokTy == cEOF {
			break
		}

		res, err := tp.rfwo(pfns)

		if err != nil {
			panic(err)
		}

		output += strings.Repeat("\t", ident) + res + "\n"

		tp.advance(1)
	}

	return output
}

// parsing functions

func (tp *Transpiler) fn() (string, error) {
	// "." id "(" "|" (id ("," id)*)? "|" code ")"

	var fname string
	var args string
	var code string

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

	for tok.tokTy == cIdentifier {
		args += tok.lexeme

		tp.advance(1)

		if tp.ctoken().tokTy != cComma {
			break
		}

		args += ","

		tp.advance(1)
		tok = tp.ctoken()
	}

	tok = tp.ctoken()

	if tok.tokTy != cBOr {
		return "", errors.New("not a function: argument list not properly closed")
	}

	code = fmt.Sprintf("def %s(%s):\n", fname, args)

	tp.advance(1)

	ident++
	code += tp.code(cRparen, tp.ret)
	ident--

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

	fmt.Println(tp.ctoken())

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

	fns := []parserFn{tp.grouping, tp.call, tp.list, tp.literal}
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
	// "<" (expr ("," expr)*)? ">"

	var output string

	tok := tp.ctoken()

	if tok.tokTy != cLt {
		return "", errors.New("not a list: no opening <")
	}

	tp.advance(1)
	tok = tp.ctoken()

	output += "["

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

	output += "]"

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
	// id "(" (expr ("," expr)*)? ")"

	var isOp bool
	var output string

	tok := tp.ctoken()

	switch tok.tokTy {
	case cIdentifier:
		isOp = false

	case cPlus:
		fallthrough
	case cMinus:
		fallthrough
	case cStar:
		fallthrough
	case cSlash:
		isOp = true

	default:
		return "", errors.New("not a function call: no operator")
	}

	if isOp {
		op := tok.lexeme
		tp.advance(1)
		tok = tp.ctoken()

		/*if tok.tokTy != cLparen {
			return "", errors.New("not a function call: missing opening paren")
		}*/

		//tp.advance(1)
		//tok = tp.ctoken()

		//if tok.tokTy != cRparen {

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

			if tok.tokTy != cComma {
				break
			}

			output += op

			tp.advance(1)
		}

		if argcount < 2 {
			return "", errors.New("not a function call: not enough arguments to function " + op)
		}

		/*if tok.tokTy != cRparen {
			return "", errors.New("not a function call: missing closing paren")
		}*/

		return output, nil
		//}

		//return "", errors.New("not a function call: not enough arguments to function " + op)
	}

	output += tok.lexeme

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cLparen {
		return "", errors.New("not a function call: missing opening paren")
	}

	output += "("

	tp.advance(1)
	tok = tp.ctoken()

	if tok.tokTy != cRparen {
		for {
			expr, err := tp.expr()

			if err != nil {
				return "", errors.New("not a function call: error parsing arguments")
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

	if tok.tokTy != cRparen {
		return "", errors.New("not a function call: missing closing )")
	}

	output += ")"

	fmt.Println(tp.ctoken())

	return output, nil
}

func (tp *Transpiler) grouping() (string, error) {
	// "(" expr ")"

	var output string

	tok := tp.ctoken()

	if tok.tokTy != cLparen {
		return "", errors.New("not a parenthesized expr at all, it even misses the opening paren")
	}

	tp.advance(1)

	output += "("
	expr, err := tp.expr()

	if err != nil {
		return "", errors.New("on grouping: error parsing expr")
	}

	tok = tp.ctoken()

	if tok.tokTy != cRparen {
		return "", errors.New("not a parenthesized expr: no matching )")
	}

	output += expr + ")"

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
	if !haderror {
		haderror = true
		fmt.Printf("had error on %s, line %d, col %d\n%s\n",
			where, tp.ctoken().line+1, tp.ctoken().col+1, msg)
	}
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
