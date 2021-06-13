package main

import (
	"strconv"
)

const (
	cLparen = iota
	cRparen
	cLbrace
	cRbrace
	cComma
	cDot
	cMinus
	cPlus
	cSemicolon
	cSlash
	cStar
	cBang
	cBangEq
	cEq
	cDoubleEq
	cGt
	cGtEq
	cLt
	cLtEq
	cIdentifier
	cString
	cArrow
	cNumber
	cBAnd
	cBOr
	cAnd
	cOr
	cEOF
	cAssignment
	cErr
	cHat
	cQuestion
	cColon
	cAt
)

func tern(cond bool, op int, op2 int) int {
	if cond {
		return op
	} // else

	return op2
}

// Token stuff {

type Token struct {
	tokTy   int
	lexeme  string
	literal interface{}
	line    int
	col     int
}

/*DEBUG: func (s Token) str() string {
	return fmt.Sprintf("%#v %#v %#v", s.tokTy, s.lexeme, s.literal)
}*/

// } token

// scanner stuff {

type Scanner struct {
	text    string
	start   int
	current int
	line    int
}

func (s *Scanner) scanTokens() []Token {
	var tokens []Token
	for !s.isAtEnd() {
		s.start = s.current
		tokens = append(tokens, s.scanToken())
	}

	tokens = append(tokens, Token{cEOF, "", nil, s.line, s.current})
	return tokens
}

func (s Scanner) isAtEnd() bool {
	return s.current >= len(s.text)
}

func (s *Scanner) scanToken() Token {
begin:
	var c byte
	if !s.isAtEnd() {
		c = s.advance()
	} else {
		return s.partialTok(cEOF)
	}

	switch c {
	case '(':
		return s.partialTok(cLparen)
	case ')':
		return s.partialTok(cRparen)
	case '{':
		return s.partialTok(cLbrace)
	case '}':
		return s.partialTok(cRbrace)
	case ',':
		return s.partialTok(cComma)
	case '.':
		return s.partialTok(cDot)
	case '-':
		return s.partialTok(tern(s.match('>'), cArrow, cMinus))
	case '+':
		return s.partialTok(cPlus)
	case ';':
		return s.partialTok(cSemicolon)
	case '*':
		return s.partialTok(cStar)
	case '?':
		return s.partialTok(cQuestion)
	case '@':
		return s.partialTok(cAt)
	case ':':
		return s.partialTok(tern(s.match('='), cAssignment, cColon))
	case '^':
		return s.partialTok(cHat)
	case '&':
		return s.partialTok(tern(s.match('&'), cAnd, cBAnd))
	case '|':
		return s.partialTok(tern(s.match('|'), cOr, cBOr))
	case '>':
		return s.partialTok(tern(s.match('='), cGtEq, cGt))
	case '<':
		return s.partialTok(tern(s.match('='), cLtEq, cLt))
	case '!':
		return s.partialTok(tern(s.match('='), cBangEq, cBang))
	case '=':
		return s.partialTok(tern(s.match('='), cDoubleEq, cEq))
	case '/':
		if !s.match('/') {
			return s.partialTok(cSlash)
		}

		for s.peek() != '\n' && !s.isAtEnd() {
			s.advance()
		}

		s.start = s.current
		goto begin

	case '"':
		return s.getstr()
	case '\n':
		s.line++
		fallthrough
	case ' ':
		fallthrough
	case '\r':
		fallthrough
	case '\t':
		s.start = s.current
		goto begin
	default:
		if s.isDigit(c) {
			return s.number()
		} else if s.isAlpha(c) {
			return s.identifier()
		} else {
			fail(s.line, "unexpected character.")
		}
	}

	return s.partialTok(cErr)
}

func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() {
		return false
	}
	if s.text[s.current] != expected {
		return false
	}
	s.current++
	return true
}

func (s Scanner) partialTok(tokTy int) Token {
	return s.mkTok(tokTy, nil)
}

func (s Scanner) mkTok(tokTy int, literal interface{}) Token {
	txt := s.text[s.start:s.current]
	return Token{tokTy, txt, literal, s.line, s.current}
}

func (s *Scanner) advance() byte {
	ret := s.text[s.current]
	s.current++
	return ret
}

func (s Scanner) peek() byte {
	if s.isAtEnd() {
		return 0
	}
	return s.text[s.current]
}

func (s *Scanner) getstr() Token {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		fail(s.line, "untermitated string")
		return s.partialTok(cEOF)
	}

	s.advance()

	value := s.text[s.start+1 : s.current-1]
	return s.mkTok(cString, value)
}

func (s Scanner) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) number() Token {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' {
		s.advance()
		if s.isDigit(s.peek()) {
			for s.isDigit(s.peek()) {
				s.advance()
			}
		}
	}

	num, _ := strconv.ParseFloat(s.text[s.start:s.current], 64)
	return s.mkTok(cNumber, num)
}

func (s Scanner) isAlpha(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		c == '_'
}

func (s *Scanner) identifier() Token {
	for s.isAlpha(s.peek()) || s.isDigit(s.peek()) {
		s.advance()
	}

	ty := cIdentifier

	return s.partialTok(ty)
}

// } scanner

func fail(line int, err string) {

}
