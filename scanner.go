package pfn

import (
	"strconv"
	"strings"
)

const (
	cLparen = iota
	cRparen
	cLbrace
	cSPC
	cRbrace
	cComma
	cEloop
	cDot
	cMinus
	cPlus
	cLAnd
	cLOr
	cSemicolon
	cSlash
	cStar
	cBang
	cDol
	cBangEq
	cT
	cEq
	cDoubleEq
	cGt
	cGtEq
	cLt
	cLtEq
	cIdentifier
	cString
	cArrow
	cMinusEq
	cPlusEq
	cStarEq
	cSlashEq
	cNumber
	cEnd
	cBAnd
	cBOr
	cAnd
	cOr
	cEOF
	cQuot
	cEqArrow
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
	kpunct  bool
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
		return s.getstr('}')
	case ',':
		return s.partialTok(cComma)
	case '.':
		return s.partialTok(cDot)
	case '-':
		return s.partialTok(tern(s.match('>'), cArrow, tern(s.match('='), cMinusEq, cMinus)))
	case '+':
		return s.partialTok(tern(s.match('='), cPlusEq, cPlus))
	case ';':
		return s.partialTok(cSemicolon)
	case '*':
		return s.partialTok(tern(s.match('='), cStarEq, cStar))
	case '?':
		return s.partialTok(cQuestion)
	case '@':
		return s.partialTok(cAt)
	case ':':
		return s.partialTok(tern(s.match('='), cAssignment, cColon))
	case '^':
		return s.partialTok(cHat)
	case '&':
		return s.partialTok(cBAnd)
	case '$':
		return s.partialTok(cDol)
	case '|':
		return s.partialTok(cBOr)
	case '~':
		return s.partialTok(cT)
	case '>':
		return s.partialTok(tern(s.match('='), cGtEq, cGt))
	case '<':
		return s.partialTok(tern(s.match('='), cLtEq, cLt))
	case '!':
		return s.partialTok(tern(s.match('='), cBangEq, cBang))
	case '=':
		if s.match('=') {
			return s.partialTok(cDoubleEq)
		}
		if s.match('>') {
			return s.partialTok(cEqArrow)
		}
		return s.partialTok(cEq)
	case '/':
		return s.partialTok(tern(s.match('='), cSlashEq, cSlash))
	case '"':
		return s.getstr('"')
	case 'f':
		if s.match('"') {
			return s.getstr('"')
		}
		return s.identifier()
	case '\'':
		return s.partialTok(cQuot)
	case '#':
		for s.peek() != '\n' && !s.isAtEnd() {
			s.advance()
		}
		s.start = s.current
		goto begin
	case ' ':
		if !s.kpunct {
			s.start = s.current
			goto begin
		}
		s.partialTok(cSPC)
	case '\n':
		s.line++
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

func (s *Scanner) getstr(end byte) Token {
	for s.peek() != end && !s.isAtEnd() {
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
	for s.isAlpha(s.peek()) || s.isDigit(s.peek()) || s.peek() == '.' || s.peek() == '_' || s.peek() == '!' || s.peek() == '@' {
		s.advance()
	}

	ty := cIdentifier

	txt := s.text[s.start:s.current]

	if strings.Contains(txt, "@") {
		return Token{cIdentifier, strings.ReplaceAll(txt, "@", "at"), nil, s.line, s.current}
	}

	if txt == "end" {
		return s.partialTok(cEnd)
	}

	if txt == "where" || txt == "while" {
		return s.partialTok(cEloop)
	}

	if txt == "and" {
		return s.partialTok(cLAnd)
	}

	if txt == "or" {
		return s.partialTok(cLOr)
	}

	return s.partialTok(ty)
}

// } scanner

func fail(line int, err string) {

}
