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
	Fragment
	Headers
	Body
)

type Scanner struct {
	src []byte // source

	mode Mode // mode (Query/Headers/Body)

	// scanning state
	ch       rune // current character
	offset   int  // character offset
	rdOffset int  // reading offset (position after current character)

	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

func New() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Init(src []byte, mode Mode) {
	s.src = src
	s.mode = mode
	s.offset = 0
	s.rdOffset = 0
	s.next()
}

func (s *Scanner) Scan() (pos injast.Pos, tok injast.Token, lit string) {
	pos = injast.Pos(s.offset)

	switch s.mode {
	case Path:
		switch s.ch {
		case '/':
			tok = injast.SLASH
		case '?':
			tok = injast.QUESTION
			s.mode = Query
		case '#':
			tok = injast.HASH
			s.mode = Fragment
		case ';':
			tok = injast.SEMICOLON
		case -1:
			tok = injast.EOF
		default:
			tok, lit = injast.IDENT, s.scanLiteral()
		}
		if tok != injast.IDENT {
			s.next()
		}
		return pos, tok, lit
	case Query:
		switch s.ch {
		case -1:
			tok = injast.EOF
		case '&':
			tok = injast.AND
		case '=':
			tok = injast.ASSIGN
		case '#':
			tok = injast.HASH
			s.mode = Fragment
		case '[':
			tok = injast.LBRACK
		case ']':
			tok = injast.RBRACK
		default:
			tok, lit = injast.IDENT, s.scanLiteral()
		}
		if tok != injast.IDENT {
			s.next()
		}
		return pos, tok, lit
	}

	return pos, tok, lit
}

func (s *Scanner) scanLiteral() string {
	offs := s.offset
	switch s.mode {
	case Path:
		for !isPathToken(s.ch) && s.ch != -1 {
			s.next()
		}
	case Query:
		for !isQueryToken(s.ch) && s.ch != -1 {
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
	return ch == '/' || ch == '?' || ch == ';' || ch == '#'
}

func isQueryToken(ch rune) bool {
	return ch == '=' || ch == '&'
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
