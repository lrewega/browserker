package parsers

import (
	"go/token"

	"gitlab.com/browserker/scanner/injections/injast"
)

type queryParser struct {
	// Next token
	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal
}

func Parse(query string) *injast.Node {
	return nil
}
