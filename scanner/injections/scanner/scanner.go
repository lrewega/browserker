package scanner

import (
	"unicode"
	"unicode/utf8"

	"gitlab.com/browserker/scanner/injections/injast"
)

type Mode int

const (
	Path Mode = iota
	Query
	Headers
	Body
)

type Scanner struct {
	src []byte // source

	Mode Mode // mode (Query/Headers/Body)

	// scanning state
	ch       rune // current character
	offset   int  // character offset
	rdOffset int  // reading offset (position after current character)

	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

func New(src []byte) *Scanner {
	return &Scanner{src: src}
}

func (s *Scanner) Scan() (pos injast.Pos, tok injast.Token, lit string) {
	pos = injast.Pos(s.offset)
	switch ch := s.ch; {
	case isLetter(ch):
		lit = s.scanLiteral()
	case ch == '.' && rune(s.peek()) == '.':
		tok, lit = injast.DOTDOT, ".."
	default:
		s.next() // always make progress
		switch ch {
		case '?':
			tok = injast.QUESTION
			lit = s.scanLiteral()
		case '#':
			tok = injast.HASH
		case '/':
			tok = injast.SLASH
		case '=':
			tok = injast.ASSIGN
		case '[':
			tok = injast.LBRACK
		case ']':
			tok = injast.RBRACK
		case '{':
			tok = injast.LBRACE
		case '}':
			tok = injast.RBRACE
		case '<':
			tok = injast.LSS
		case '>':
			tok = injast.GTR
		case ':':
			tok = injast.COLON
		case '\'':
			tok = injast.SQUOTE
		case '"':
			tok = injast.DQUOTE
		case '&':
			tok = injast.DQUOTE
		case ',':
			tok = injast.COMMA
		}
	}
	return pos, tok, lit
}

func (s *Scanner) scanLiteral() string {
	offs := s.offset
	switch s.Mode {
	case Path:
		for !isPathToken(s.ch) {
			s.next()
		}
	default:
		for isLetter(s.ch) || isDigit(s.ch) {
			s.next()
		}
	}

	return string(s.src[offs:s.offset])
}

// peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, peek returns 0.
func (s *Scanner) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		s.ch = -1 // eof
	}
}

// Denotes the end of a path
func isPathToken(ch rune) bool {
	return ch == '/' || ch == '?' || ch == ';'
}

func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }
