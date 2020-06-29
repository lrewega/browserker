package scanner

import (
	"unicode"
	"unicode/utf8"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/token"
)

// Mode of scanner (determines end of mode tokens)
type Mode int

const (
	URI Mode = iota
	Headers
	Cookies
	Body
	MultipartBody
)

type Scanner struct {
	src []byte // source

	mode Mode // mode (Query/Headers/Body)

	// scanning state
	ch       rune // current character
	offset   int  // character offset
	rdOffset int  // reading offset (position after current character)

}

// New token scanner
func New() *Scanner {
	return &Scanner{}
}

// Mode that the scanner is currently in
func (s *Scanner) Mode() Mode {
	return s.mode
}

// Init the scanner with source and starting mode (Note: Mode may change during parsing)
func (s *Scanner) Init(src []byte, mode Mode) {
	s.src = src
	s.mode = mode
	s.offset = 0
	s.rdOffset = 0
	s.next()
}

func (s *Scanner) Scan() (pos browserk.InjectionPos, tok token.Token, lit string) {
	s.skipWhitespace()
	pos = browserk.InjectionPos(s.offset)

	switch s.mode {
	case URI:
		tok, lit = s.scanURI()
	case Cookies:
		tok, lit = s.scanCookies()
	case Body:
		tok, lit = s.scanBody()
	}
	if tok != token.IDENT {
		s.next()
	}
	return pos, tok, lit
}

// scanPath for tokens or switch mode to Query or Fragment
func (s *Scanner) scanURI() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case '/':
		tok = token.SLASH
	case '?':
		tok = token.QUESTION
	case '#':
		tok = token.HASH
	case ';':
		tok = token.SEMICOLON
	case '&':
		tok = token.AND
	case '=':
		tok = token.ASSIGN
	case '[':
		tok = token.LBRACK
	case ']':
		tok = token.RBRACK
	default:
		tok, lit = token.IDENT, s.scanLiteral()
	}
	return tok, lit
}

// scanCookies for tokens
func (s *Scanner) scanCookies() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case ';':
		tok = token.SEMICOLON
	case '=':
		tok = token.ASSIGN
	case ' ':
		tok = token.SPACE
	case ',':
		tok = token.COMMA
	}
	return tok, lit
}

// scanBody for tokens (x-www-url-encoded, json, and xml)
func (s *Scanner) scanBody() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case '&':
		tok = token.AND
	case '=':
		tok = token.ASSIGN
	case '[':
		tok = token.LBRACK
	case ']':
		tok = token.RBRACK
		/*
			case '{':
				tok = token.LBRACE
			case '}':
				tok = token.LBRACK
			case '<':
				tok = token.LSS
			case '>':
				tok = token.GTR
			case ':':
				tok = token.COLON
			case '"':
				tok = token.DQUOTE
			case '\'':
				tok = token.SQUOTE
			case ',':
				tok = token.COMMA
		*/
	default:
		tok, lit = token.IDENT, s.scanLiteral()
	}
	return tok, lit
}

func (s *Scanner) scanBodyJSON() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case '{':
		tok = token.LBRACE
	case '}':
		tok = token.RBRACE
	case '[':
		tok = token.LBRACK
	case ']':
		tok = token.RBRACK
	case ':':
		tok = token.COLON
	case ',':
		tok = token.COMMA
	case '\'':
		tok = token.SQUOTE
	case '"':
		tok = token.DQUOTE
	default:
		tok, lit = token.IDENT, s.scanLiteral()
	}
	return tok, lit
}

func (s *Scanner) scanBodyXML() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case '<':
		tok = token.LSS
	case '>':
		tok = token.GTR
	case ':':
		tok = token.COLON
	case '\'':
		tok = token.SQUOTE
	case '"':
		tok = token.DQUOTE
	case '=':
		tok = token.ASSIGN
	case '/':
		tok = token.SLASH
	default:
		tok, lit = token.IDENT, s.scanLiteral()
	}
	return tok, lit
}

// PeekIsBodyJSON Checks if we are in a JSON object/array definition
func (s *Scanner) PeekIsBodyJSON() bool {
	var open rune
	// if we don't start with an open { or [ it's probably not JSON.
	// TODO: ack, what if we are in bodyXML? (json in xml) need to check history
	if s.offset != 0 && s.PeekBackwards() != '=' {
		return false
	}

	if s.PeekBackwards() == '{' {
		open = '{'
	}
	if s.PeekBackwards() == '[' {
		open = '['
	}

	// Not JSON
	if open == rune(0) {
		return false
	}

	rdOffset := s.rdOffset
	trackOpen := 1
	inQuotes := 0 // for making sure we don't count { or [ if we are inside of quotes

	for ; s.rdOffset < len(s.src); s.rdOffset++ {
		switch s.Peek() {
		case '{':
			if '{' == open && (inQuotes%2 == 0) {
				trackOpen++
			}
		case '}':
			if '{' == open && (inQuotes%2 == 0) {
				trackOpen--
			}
		case '[':
			if '[' == open && (inQuotes%2 == 0) {
				trackOpen++
			}
		case ']':
			if '[' == open && (inQuotes%2 == 0) {
				trackOpen--
			}
		case '\'':
			inQuotes++
		case '"':
			inQuotes++
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
	// nope just random { or ['s i guess
	s.rdOffset = rdOffset

	if trackOpen == 0 {
		return true
	}
	return false
}

// PeekIsBodyXML checks if we are XML
func (s *Scanner) PeekIsBodyXML() bool {
	var open rune
	// if we don't start with an open { or [ it's probably not JSON.
	// TODO: ack, what if we are in bodyXML? (json in xml) need to check history
	if s.offset != 0 && s.PeekBackwards() != '=' {
		return false
	}

	if s.PeekBackwards() == '<' {
		open = '<'
	}

	// Not XML
	if open == rune(0) {
		return false
	}

	rdOffset := s.rdOffset
	trackOpen := 1
	inQuotes := 0 // for making sure we don't count { or [ if we are inside of quotes

	for ; s.rdOffset < len(s.src); s.rdOffset++ {
		switch s.Peek() {
		case '<':
			if inQuotes%2 == 0 {
				trackOpen++
			}
		case '>':
			if inQuotes%2 == 0 {
				trackOpen--
			}
		case '\'':
			inQuotes++
		case '"':
			inQuotes++
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
	// nope just random <'s i guess
	s.rdOffset = rdOffset

	if trackOpen == 0 {
		return true
	}
	return false
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' {
		s.next()
	}
}

func (s *Scanner) scanLiteral() string {
	offs := s.offset
	switch s.mode {
	case URI:
		for !isURIToken(s.ch) && s.ch != -1 {
			s.next()
		}
	case Cookies:
		for !isCookieToken(s.ch) && s.ch != -1 {
			s.next()
		}
	case Body:
		for !isBody(s.ch) && s.ch != -1 {
			s.next()
		}
	default:
		for isLetter(s.ch) || isDigit(s.ch) {
			s.next()
		}
	}

	return string(s.src[offs:s.offset])
}

// Peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, Peek returns 0.
func (s *Scanner) Peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

// PeekBackwards returns the byte prior to the most recently read character without
// advancing the scanner. If the scanner is already at the first byte return 0
func (s *Scanner) PeekBackwards() byte {
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
func isURIToken(ch rune) bool {
	return ch == '/' || ch == '?' || ch == ';' || ch == '#' || isQueryToken(ch)
}

func isQueryToken(ch rune) bool {
	return ch == '=' || ch == '&' || ch == '[' || ch == ']' || ch == '#'
}

func isCookieToken(ch rune) bool {
	return ch == ';' || ch == '=' || ch == ' ' || ch == ','
}

func isBody(ch rune) bool {
	return ch == '&' || ch == '=' || isBodyJSON(ch) || isBodyXML(ch)
}

func isBodyJSON(ch rune) bool {
	return ch == '{' || ch == '}' || ch == '[' || ch == ']' || ch == ':' || ch == ',' || ch == '\'' || ch == '"'
}

func isBodyXML(ch rune) bool {
	return ch == '<' || ch == '>' || ch == ':' || ch == '\'' || ch == '"' || ch == '=' || ch == '/'
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
