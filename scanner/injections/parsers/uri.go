package parsers

import (
	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/scanner"
	"gitlab.com/browserker/scanner/injections/token"
)

type QueryMode int

const (
	Path QueryMode = iota + 1
	File
	Query
	Fragment
	FragmentPath
	FragmentQuery
)

type keyValueMode int

const (
	keyMode keyValueMode = iota + 1
	valueMode
)

type URIParser struct {
	s       *scanner.Scanner
	mode    QueryMode
	kvMode  keyValueMode
	kvIndex int
	uri     *injast.URI
}

func (u *URIParser) Parse(uri string) (*injast.URI, error) {
	u.s = scanner.New()
	u.s.Init([]byte(uri), scanner.URI)
	u.uri = injast.NewURI([]byte(uri))
	u.mode = Path
	u.kvMode = keyMode
	u.kvIndex = 0

	for {
		pos, tok, lit := u.s.Scan()
		if tok == token.EOF {
			return u.uri, nil
		}
		u.updateMode(tok)

		switch u.mode {
		case Path:
			if tok == token.SLASH {
				continue
			} else if tok.IsLiteral() {
				// peek backwards because the scanner is already primed for the next token
				// calling peek would skip past the potential end of path/file delimiter
				peek := u.s.PeekBackwards()

				// if the next char is a ?, ;, # or EOF that means this ident is a file part
				if peek == '?' || peek == '#' || peek == ';' || peek == 0 {
					u.uri.File = &injast.Ident{
						NamePos: pos,
						Name:    lit,
					}
				} else {
					u.uri.Paths = append(u.uri.Paths, &injast.Ident{
						NamePos: pos,
						Name:    lit,
					})
				}
			}
		// case File: file is a one shot, added under case Path so no need to capture here
		case Query:
			u.handleParams(tok, pos, lit, &u.uri.Query.Params)
		case Fragment:
			switch tok {
			case token.HASH:
				continue
			case token.SLASH:
				u.mode = FragmentPath
				continue
			case token.IDENT:
				peek := u.s.Peek()
				if peek == '?' || peek == '&' || peek == 0 {
					u.uri.Fragment.File = &injast.Ident{
						NamePos: pos,
						Name:    lit,
					}
					continue
				}
				//prevTok := u.s.PeekBackwards()
				u.mode = FragmentQuery
				u.handleParams(tok, pos, lit, &u.uri.Fragment.Params)
			}
		case FragmentPath:
			if tok == token.SLASH {
				continue
			} else if tok.IsLiteral() {
				peek := u.s.Peek()

				// if the next char is a ? & or EOF that means this ident is a file part
				if peek == '?' || peek == '&' || peek == 0 {
					u.uri.Fragment.File = &injast.Ident{
						NamePos: pos,
						Name:    lit,
					}
				} else {
					u.uri.Fragment.Paths = append(u.uri.Fragment.Paths, &injast.Ident{
						NamePos: pos,
						Name:    lit,
					})
				}
			}
		case FragmentQuery:
			u.handleParams(tok, pos, lit, &u.uri.Fragment.Params)
		}
	}
}

func (u *URIParser) handleParams(tok token.Token, pos injast.Pos, lit string, params *[]*injast.KeyValueExpr) {
	switch tok {
	case token.ASSIGN:
		// /file?=asdf (invalid, but we must account for it)
		if len(*params) == 0 {
			*params = append(*params, &injast.KeyValueExpr{
				Key:     &injast.Ident{NamePos: pos, Name: lit},
				Sep:     0,
				SepChar: 0,
				Value:   nil,
			})
		}
		(*params)[u.kvIndex].Sep = pos
		(*params)[u.kvIndex].SepChar = '='
		u.kvMode = valueMode
		return
	case token.AND:
		u.kvIndex++
		u.kvMode = keyMode
		return
	case token.QUESTION:
		return
	}

	if u.kvMode == keyMode {
		var key injast.Expr

		key = &injast.Ident{NamePos: pos, Name: lit}
		peek := u.s.PeekBackwards()
		if peek == '[' {
			key = u.handleIndexExpr(pos, lit)
		}
		*params = append(*params, &injast.KeyValueExpr{
			Key:     key,
			Sep:     0,
			SepChar: 0,
			Value:   nil,
		})
	} else {
		(*params)[u.kvIndex].Value = &injast.Ident{NamePos: pos, Name: lit}
	}
}

func (u *URIParser) handleIndexExpr(originalPos injast.Pos, lit string) injast.Expr {
	expr := &injast.IndexExpr{
		X:      &injast.Ident{NamePos: originalPos, Name: lit},
		Lbrack: 0,
		Index:  nil,
		Rbrack: 0,
	}

	notIndexExpr := lit

	for {
		pos, tok, lit := u.s.Scan()
		switch tok {
		case token.EOF, token.ASSIGN:
			return &injast.Ident{NamePos: originalPos, Name: notIndexExpr}
		case token.LBRACK:
			notIndexExpr += "["
			expr.Lbrack = pos
		case token.RBRACK:
			notIndexExpr += "]"
			expr.Rbrack = pos
			return expr
		default:
			expr.Index = &injast.Ident{NamePos: pos, Name: lit}
			notIndexExpr += lit
		}
	}
}

func (u *URIParser) updateMode(tok token.Token) {
	if (u.mode == Path && tok == token.QUESTION) ||
		(u.mode == File && tok == token.QUESTION) {
		u.mode = Query
	} else if (u.mode == Path && tok == token.HASH) ||
		(u.mode == File && tok == token.HASH) ||
		(u.mode == Query && tok == token.HASH) {
		u.mode = Fragment
	} else if u.mode == Fragment && tok == token.SLASH {
		u.mode = FragmentPath
	} else if u.mode == FragmentPath && tok == token.QUESTION {
		u.mode = FragmentQuery
	}
}
