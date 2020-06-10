package scanner

import (
	"unicode"
	"unicode/utf8"

	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/token"
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

// Mode that the scanner is currently in
func (s *Scanner) Mode() Mode {
	return s.mode
}

// ModeHistory of the scanner
func (s *Scanner) ModeHistory() []Mode {
	return s.modeHistory
}

// Init the scanner with source and starting mode (Note: Mode may change during parsing)
func (s *Scanner) Init(src []byte, mode Mode) {
	s.src = src
	s.mode = mode
	s.offset = 0
	s.rdOffset = 0
	s.next()
}

func (s *Scanner) Scan() (pos injast.Pos, tok token.Token, lit string) {
	s.skipWhitespace()
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
		tok, lit = s.scanBodyJSON()
	case BodyXML:
		tok, lit = s.scanBodyXML()
	}
	if tok != token.IDENT {
		s.next()
	}
	return pos, tok, lit
}

// scanPath for tokens or switch mode to Query or Fragment
func (s *Scanner) scanPath() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case '/':
		tok = token.SLASH
	case '?':
		tok = token.QUESTION
		s.modeHistory = append(s.modeHistory, Path)
		s.mode = Query
	case '#':
		tok = token.HASH
		s.modeHistory = append(s.modeHistory, Path)
		s.mode = Fragment
	case ';':
		tok = token.SEMICOLON
	default:
		tok, lit = token.IDENT, s.scanLiteral()
	}
	return tok, lit
}

// scanQuery for tokens or switch mode to Fragment
func (s *Scanner) scanQuery() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case '&':
		tok = token.AND
	case '=':
		tok = token.ASSIGN
	case '#':
		tok = token.HASH
		s.modeHistory = append(s.modeHistory, Query)
		s.mode = Fragment
	case '[':
		tok = token.LBRACK
	case ']':
		tok = token.RBRACK
	default:
		tok, lit = token.IDENT, s.scanLiteral()
	}
	return tok, lit
}

// scanFragment acts much like scanQuery since it's up to the developer
// how they structure their app (some use it like a path, some like a query)
func (s *Scanner) scanFragment() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case '/':
		tok = token.SLASH
	case '&':
		tok = token.AND
	case '=':
		tok = token.ASSIGN
	case '#':
		tok = token.HASH
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

// scanBody for tokens and sniff for JSON or XML and update mode accordingly
func (s *Scanner) scanBody() (tok token.Token, lit string) {
	switch s.ch {
	case -1:
		tok = token.EOF
	case '&':
		tok = token.AND
	case '=':
		tok = token.ASSIGN
	case '[':
		if s.scanIsJSON('[') {
			if s.offset != 0 {
				// only change history if we were in a x-www-url-encoded body first
				// if we are the first byte then the entire request is probably JSON
				s.modeHistory = append(s.modeHistory, s.mode)
			}
			s.mode = BodyJSON
		}
		tok = token.LBRACK
	case ']':
		tok = token.RBRACK
	case '<':
		if s.scanIsXML() {
			if s.offset != 0 {
				s.modeHistory = append(s.modeHistory, s.mode)
			}
			s.mode = BodyXML
			tok = token.LSS
		} else {
			tok, lit = token.IDENT, s.scanLiteral()
		}
	case '{':
		if s.scanIsJSON('{') {
			if s.offset != 0 {
				// only change history if we were in a x-www-url-encoded body first
				// if we are the first byte then the entire request is probably JSON
				s.modeHistory = append(s.modeHistory, s.mode)
			}
			s.mode = BodyJSON
			tok = token.LBRACE
		} else {
			tok, lit = token.IDENT, s.scanLiteral()
		}
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

// Checks if we are in a JSON object/array definition
func (s *Scanner) scanIsJSON(open rune) bool {
	// if we don't start with an open { or [ it's probably not JSON.
	// TODO: ack, what if we are in bodyXML? (json in xml) need to check history
	if s.offset != 0 && s.peekBackwards() != '=' {
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

	if trackOpen == 0 {
		return true
	}
	return false
}

func (s *Scanner) scanIsXML() bool {
	// if we don't start with an open { or [ it's probably not JSON.
	// TODO: ack, what if we are in bodyXML? (json in xml) need to check history
	if s.offset != 0 && s.peekBackwards() != '=' {
		return false
	}
	rdOffset := s.rdOffset
	trackOpen := 1
	for ; s.rdOffset < len(s.src); s.rdOffset++ {
		switch s.peek() {
		case '<':
			trackOpen++
		case '>':
			trackOpen--
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

	if trackOpen == 0 {
		return true
	}
	// nope just random <'s i guess
	s.rdOffset = rdOffset
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
	case Body:
		for !isBody(s.ch) && s.ch != -1 {
			s.next()
		}
	case BodyJSON:
		for !isBodyJSON(s.ch) && s.ch != -1 {
			s.next()
		}
	case BodyXML:
		for !isBodyXML(s.ch) && s.ch != -1 {
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

func isBody(ch rune) bool {
	return ch == '&' || ch == '='
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
