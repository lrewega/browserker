package parsers

import (
	"github.com/browserker/scanner/injections/token"
	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/scanner"
)

func Parse(uri string) (*injast.URI, error) {
	s := scanner.New()
	s.Init([]byte(uri), scanner.Path)
	u := injast.NewURI()
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			return u, nil
		}
		switch s.Mode() {
		case scanner.Path:
			if tok == token.SLASH {
				continue
			} else if tok.IsLiteral() {
				u.Paths = append(u.Paths, &injast.Ident{
					NamePos: pos,
					Name:    lit,
				})
			}
		case scanner.Query:
		case scanner.Fragment:
		}
	}

	return u, nil
}
