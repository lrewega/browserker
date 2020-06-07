package injast

import (
	"strconv"
)

// Token is the set of lexical tokens of the Go programming language.
type Token int

// The list of tokens.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF

	literal_beg
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENT         // main
	INT           // 12345
	FLOAT         // 123.45
	IMAG          // 123.45i
	DOTDOT        // ..
	SQUOTE_STRING // 'a'
	DQUOTE_STRING // "abc"
	literal_end

	delim_beg
	ASSIGN    // =
	SLASH     // /
	AND       // &
	LSS       // <
	GTR       // >
	LPAREN    // (
	LBRACK    // [
	LBRACE    // {
	COMMA     // ,
	PERIOD    // .
	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
	SQUOTE    // '
	DQUOTE    // "
	BACKTICK  // `
	QUESTION  // ?
	HASH      // #
	delim_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	IDENT:         "IDENT",
	INT:           "INT",
	FLOAT:         "FLOAT",
	IMAG:          "IMAG",
	SQUOTE_STRING: "SQUOTE_STRING",
	DQUOTE_STRING: "DQUOTE_STRING",

	ASSIGN:    "=",
	SLASH:     "/",
	AND:       "&",
	LSS:       "<",
	GTR:       ">",
	LPAREN:    "(",
	LBRACK:    "[",
	LBRACE:    "{",
	COMMA:     ",",
	PERIOD:    ".",
	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",
	SQUOTE:    "'",
	DQUOTE:    `"`,
	BACKTICK:  "`",
	QUESTION:  "?",
	HASH:      "#",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
//
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
//
func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

// IsDelimiter returns true for tokens corresponding to
// delimiters; it returns false otherwise.
//
func (tok Token) IsDelimiter() bool { return delim_beg < tok && tok < delim_end }
