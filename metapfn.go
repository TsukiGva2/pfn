package main

import (
	"errors"
	"fmt"
)

type macro struct {
	args []string
	fmt  string
}

type meta struct {
	output  string
	tokens  []Token
	current uint
	macros  map[string]macro
}

type metaFn (func() (string, error))

func runMeta(code string) string {
	sc := Scanner{code, 0, 0, 0, true}
	tokens := sc.scanTokens()

	mt := meta{}
	mt.tokens = tokens
	mt.macros = make(map[string]macro)

	mt.code()

	fmt.Println(mt.output)

	return mt.output
}

func (mt *meta) code() {
	line := 0
	for {
		tok := mt.ctoken()

		if tok.tokTy == cEOF {
			break
		}

		err := mt.wfwo(mt.macro, mt.call)

		if err != nil {
			mt.output += tok.lexeme

			if tok.line > line {
				mt.output += "\n"
				line = tok.line
			}
		}

		mt.advance()
	}
}

func (mt *meta) wfwo(fns ...metaFn) error {
	for i := range fns {
		old := mt.current
		res, err := fns[i]()

		if err != nil {
			mt.current = old
			continue
		}

		mt.output += res + "\n"
		return nil
	}

	return errors.New("no working function")
}

func (mt *meta) call() (string, error) {
	//var re = regexp.MustCompile(`(^|[^_])\bproducts\b([^_]|$)`)
	//s := re.ReplaceAllString(sample, `$1.$2`)

	// "(" id

	return "", errors.New("not implemented yet")
}

func (mt *meta) macro() (string, error) {
	// "$" "." id "(" "|" args "|" code ")"
	var output string
	var mname string
	var args []string
	var body string

	mt.skipSpaces()

	tok := mt.ctoken()

	if tok.tokTy != cDol {
		return "", errors.New("not a macro definition: no '$'")
	}

	mt.advance()
	mt.skipSpaces()
	tok = mt.ctoken()

	if tok.tokTy != cDot {
		return "", errors.New("not a macro definition: no '.'")
	}

	mt.advance()
	mt.skipSpaces()
	tok = mt.ctoken()

	if tok.tokTy != cIdentifier {
		return "", errors.New("not a macro definition: no identifier")
	}

	mname = tok.lexeme

	mt.advance()
	mt.skipSpaces()
	tok = mt.ctoken()

	if tok.tokTy != cLparen {
		return "", errors.New("not a macro definition: no '('")
	}

	mt.advance()
	mt.skipSpaces()
	tok = mt.ctoken()

	if tok.tokTy != cBOr {
		return "", errors.New("not a macro definition: no '|'")
	}

	mt.advance()
	mt.skipSpaces()
	tok = mt.ctoken()

	for {
		tok = mt.ctoken()

		if tok.tokTy != cIdentifier {
			return "", errors.New("not a macro definition: invalid argument name in arguments definition")
		}

		args = append(args, tok.lexeme)

		mt.advance()
		mt.skipSpaces()
		tok = mt.ctoken()

		if tok.tokTy != cComma {
			break
		}

		mt.advance()
		mt.skipSpaces()
		tok = mt.ctoken()
	}

	if tok.tokTy != cBOr {
		return "", errors.New("not a macro definition: no closing ')'")
	}

	mt.advance()
	mt.skipSpaces()
	tok = mt.ctoken()

	line := tok.line
	for {
		if tok.tokTy == cRparen || tok.tokTy == cEOF {
			break
		}

		body += tok.lexeme

		if tok.line > line {
			body += "\n"
			line = tok.line
		}

		mt.advance()
		tok = mt.ctoken()
	}

	if tok.tokTy != cRparen {
		return "", errors.New("not a macro definition: no ')'")
	}

	mt.macros[mname] = macro{args, body}

	return output, nil
}

func (mt *meta) skipSpaces() {
	for mt.ctoken().tokTy == cSPC {
		mt.advance()
	}
}

func (mt *meta) ctoken() Token {
	return mt.tokens[mt.current]
}

func (mt *meta) advance() {
	if mt.current+1 < uint(len(mt.tokens)) {
		mt.current++
	}
}
