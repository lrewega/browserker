package scanner

import (
	"unicode"
	"unicode/utf8"

	"gitlab.com/browserker/scanner/injections/injast"
)

// Mode of scanner (determines end of mode tokens)
type Mode int

const (
	Path Mode = iota
	Query
	Fragment
	Headers
	Cookies
	Body
	BodyJSON
	BodyXML
	MultipartBody
)

type Scanner struct {
	src []byte // source

	mode        Mode // mode (Query/Headers/Body)
	modeHistory []Mode

	// scanning state
	ch       rune // current character
	offset   int  // character offset
	rdOffset int  // reading offset (position after current character)

	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

// New token scanner
func New() *Scanner {
	return &Scanner{modeHistory: make([]Mode, 0)}
}

// Init the scanner with source and starting mode (Note: Mode may change during parsing)
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
		tok, lit = s.scanPath()
	case Query:
		tok, lit = s.scanQuery()
	case Fragment: // treat Fragments much like Query
		tok, lit = s.scanFragment()
	case Cookies:
		tok, lit = s.scanCookies()
	case Body:
		tok, lit = s.scanBody()
	case BodyJSON:
	case BodyXML:
	}
	if tok != injast.IDENT {
		s.next()
	}
	return pos, tok, lit
}

func (s *Scanner) scanPath() (tok injast.Token, lit string) {
	switch s.ch {
	case -1:
		tok = injast.EOF
	case '/':
		tok = injast.SLASH
	case '?':
		tok = injast.QUESTION
		s.modeHistory = append(s.modeHistory, Path)
		s.mode = Query
	case '#':
		tok = injast.HASH
		s.modeHistory = append(s.modeHistory, Path)
		s.mode = Fragment
	case ';':
		tok = injast.SEMICOLON
	default:
		tok, lit = injast.IDENT, s.scanLiteral()
	}
	return tok, lit
}

func (s *Scanner) scanQuery() (tok injast.Token, lit string) {
	switch s.ch {
	case -1:
		tok = injast.EOF
	case '&':
		tok = injast.AND
	case '=':
		tok = injast.ASSIGN
	case '#':
		tok = injast.HASH
		s.modeHistory = append(s.modeHistory, Query)
		s.mode = Fragment
	case '[':
		tok = injast.LBRACK
	case ']':
		tok = injast.RBRACK
	default:
		tok, lit = injast.IDENT, s.scanLiteral()
	}
	return tok, lit
}

func (s *Scanner) scanFragment() (tok injast.Token, lit string) {
	switch s.ch {
	case -1:
		tok = injast.EOF
	case '/':
		tok = injast.SLASH
	case '&':
		tok = injast.AND
	case '=':
		tok = injast.ASSIGN
	case '#':
		tok = injast.HASH
	case '[':
		tok = injast.LBRACK
	case ']':
		tok = injast.RBRACK
	default:
		tok, lit = injast.IDENT, s.scanLiteral()
	}
	return tok, lit
}

func (s *Scanner) scanCookies() (tok injast.Token, lit string) {
	switch s.ch {
	case -1:
		tok = injast.EOF
	case ';':
		tok = injast.SEMICOLON
	case '=':
		tok = injast.ASSIGN
	case ' ':
		tok = injast.SPACE
	case ',':
		tok = injast.COMMA
	}
	return tok, lit
}

func (s *Scanner) scanBody() (tok injast.Token, lit string) {
	switch s.ch {
	case -1:
		tok = injast.EOF
	case '&':
		tok = injast.AND
	case '=':
		tok = injast.ASSIGN
	case '[':
		if s.scanIsJSON('[') {
			s.mode = BodyJSON
			if s.offset != 0 {
				// only change history if we were in a x-www-url-encoded body first
				// if we are the first byte then the entire request is probably JSON
				s.modeHistory = append(s.modeHistory, Body)
			}
		}
		tok = injast.LBRACK
	case ']':
		tok = injast.RBRACK
	case '{':
		if s.scanIsJSON('{') {
			s.mode = BodyJSON
			tok = injast.LBRACE
			if s.offset != 0 {
				// only change history if we were in a x-www-url-encoded body first
				// if we are the first byte then the entire request is probably JSON
				s.modeHistory = append(s.modeHistory, Body)
			}
		} else {
			tok, lit = injast.IDENT, s.scanLiteral()
		}
	default:
		tok, lit = injast.IDENT, s.scanLiteral()
	}
	return tok, lit
}

// Checks if we are in a JSON object definition
func (s *Scanner) scanIsJSON(open rune) bool {
	// if we don't start with an open { or [ it's probably not JSON.
	// TODO: ack, what if we are in bodyXML? (json in xml) need to check history
	if s.peekBackwards() != '=' {
		return false
	}
	rdOffset := s.rdOffset
	trackOpen := 1
	for ; s.rdOffset < len(s.src); s.rdOffset++ {
		switch s.peek() {
		case '{':
			if '{' == open {
				trackOpen++
			}
		case '}':
			if '{' == open {
				trackOpen--
			}
		case '[':
			if '[' == open {
				trackOpen++
			}
		case ']':
			if '[' == open {
				trackOpen--
			}
		case '&':
			if trackOpen == 0 {
				s.rdOffset = rdOffset // reset the offset since we are just peeking
				return true
			}
		case 0:
			// reached EOF so this was either a full on JSON request or the last
			// body param was a JSON value (if trackOpen is 0)
			if trackOpen == 0 {
				s.rdOffset = rdOffset // reset the offset since we are just peeking
				return true
			}
			break
		}
	}
	// nope just random {'s i guess
	s.rdOffset = rdOffset
	return false
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
	case Fragment:
		for !isFragmentToken(s.ch) && s.ch != -1 {
			s.next()
		}
	case Cookies:
		for !isCookieToken(s.ch) && s.ch != -1 {
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

// peek returns the byte prior to the most recently read character without
// advancing the scanner. If the scanner is already at the first byte return 0
func (s *Scanner) peekBackwards() byte {
	if s.rdOffset-1 >= 0 {
		return s.src[s.rdOffset-1]
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
	return ch == '=' || ch == '&' || ch == '[' || ch == ']' || ch == '#'
}

func isFragmentToken(ch rune) bool {
	return isQueryToken(ch) || ch == '/'
}

func isCookieToken(ch rune) bool {
	return ch == ';' || ch == '=' || ch == ' ' || ch == ','
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
