package parsers

import (
	"fmt"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/scanner"
	"gitlab.com/browserker/scanner/injections/token"
)

type URIParser struct {
	s   *scanner.Scanner
	uri *injast.URI
}

// Parse a uri into it's parts
// TODO: Clean this up it's a disaster
func (u *URIParser) Parse(uri string) (*injast.URI, error) {
	u.s = scanner.New()
	u.s.Init([]byte(uri), scanner.URI)
	u.uri = injast.NewURI([]byte(uri))

	for {
		_, tok, lit := u.s.Scan()
		switch tok {
		case token.EOF:
			return u.uri, nil
		case token.SLASH:
			u.parsePath(browserk.InjectPath)
		default:
			return u.uri, fmt.Errorf("uri did not start with a /, got %s treating as invalid", lit)
		}
	}
}

func (u *URIParser) parsePath(paramLoc browserk.InjectionLocation) {
	path := &injast.Ident{Location: paramLoc}
	for {
		_, tok, lit := u.s.Scan()

		switch tok {
		case token.EOF:
			if path != nil {
				u.uri.Fields = append(u.uri.Fields, path)
			}
			return
		// we are a path
		case token.SLASH:
			u.uri.Fields = append(u.uri.Fields, path)
			u.parsePath(paramLoc)
			path = nil
		// we are a file, terminated by semicolon or question
		case token.SEMICOLON, token.QUESTION:
			path.Location = browserk.InjectFile
			u.uri.Fields = append(u.uri.Fields, path)
			u.parseQuery(browserk.InjectQueryName)
			path = nil
		// we are a file, terminated by #
		case token.HASH:
			path.Location = browserk.InjectFile
			u.uri.Fields = append(u.uri.Fields, path)
			u.parseFragment()
			path = nil
		default:
			path.Name += lit
		}
	}
}

func (u *URIParser) parseQuery(paramLoc browserk.InjectionLocation) {
	for {
		kv := &injast.KeyValueExpr{Location: paramLoc}
		u.uri.Fields = append(u.uri.Fields, kv)
	SCAN:
		_, tok, lit := u.s.Scan()
		switch tok {
		case token.EOF:
			return
		case token.AND:
			if paramLoc == browserk.InjectQueryValue {
				paramLoc = browserk.InjectQuery
				peek := u.s.Peek()
				// if we are end of line, we don't want to go to top of for loop and add an empty kv
				if peek == 0 || peek == '#' {
					goto SCAN
				}
			}
		case token.ASSIGN:
			if paramLoc == browserk.InjectQueryName {
				paramLoc = browserk.InjectQueryValue
				kv.SepChar = '='
				goto SCAN
			}
		case token.HASH:
			u.parseFragment()
			goto SCAN
		default:
			// if kv.Key == nil ?
			if paramLoc == browserk.InjectQuery || paramLoc == browserk.InjectQueryName {
				if u.s.PeekBackwards() == '[' {
					kv.Key = u.handleIndexExpr(lit, paramLoc)
				} else {
					kv.Key = &injast.Ident{Name: lit, Location: browserk.InjectQueryName}
				}
				paramLoc = browserk.InjectQueryName
			} else {
				kv.Value = u.handleValueExpr(tok, lit, paramLoc)
			}
			goto SCAN
		}
	}
}

// make them all KV's so we can capture sepchar and keep parsing craziness to a minimum
func (u *URIParser) parseFragment() {
	kv := &injast.KeyValueExpr{Location: browserk.InjectFragment}
	for {
		_, tok, lit := u.s.Scan()

		switch tok {
		case token.EOF:
			if kv.Key != nil {
				u.uri.Fields = append(u.uri.Fields, kv)
			}
			return
		case token.SLASH:
			kv.SepChar = '/'
			u.uri.Fields = append(u.uri.Fields, kv)
			kv = &injast.KeyValueExpr{}
		case token.AND:
			kv.SepChar = '&'
			u.uri.Fields = append(u.uri.Fields, kv)
			kv = &injast.KeyValueExpr{}
		case token.ASSIGN:
			kv.SepChar = '='
			u.uri.Fields = append(u.uri.Fields, kv)
			kv = &injast.KeyValueExpr{}
		case token.QUESTION:
			kv.SepChar = '?'
			u.uri.Fields = append(u.uri.Fields, kv)
			kv = &injast.KeyValueExpr{}
		default:
			if lit == "" {
				lit = tok.String()
			}
			if kv.Key != nil {
				k, _ := kv.Key.(*injast.Ident)
				k.Name += lit
			} else {
				kv.Key = &injast.Ident{Name: lit}
			}
		}
	}
}

func (u *URIParser) handleValueExpr(originalTok token.Token, originalLit string, paramLoc browserk.InjectionLocation) browserk.InjectionExpr {
	if originalLit == "" {
		originalLit = originalTok.String()
	}
	value := &injast.Ident{Name: originalLit, Location: paramLoc}
	for {
		// short circuit so we don't consume the name provided we don't start with a [ or ]
		if (u.s.PeekBackwards() == '&' || u.s.PeekBackwards() == '#') && (originalTok != token.LBRACK && originalTok != token.RBRACK) {
			return value
		}
		_, tok, lit := u.s.Scan()
		switch tok {
		case token.AND, token.EOF:
			return value
		case token.LBRACK:
			value.Name += "["
		case token.RBRACK:
			value.Name += "]"
		default:
			value.Name += lit
		}
	}
}

func (u *URIParser) handleIndexExpr(lit string, paramLoc browserk.InjectionLocation) browserk.InjectionExpr {
	expr := &injast.IndexExpr{
		X:        &injast.Ident{Name: lit, Location: paramLoc},
		Index:    &injast.Ident{},
		Location: paramLoc,
	}

	notIndexExpr := lit

	for {
		_, tok, lit := u.s.Scan()
		switch tok {
		case token.EOF, token.ASSIGN:
			return &injast.Ident{Name: notIndexExpr, Location: paramLoc}
		case token.LBRACK:
			notIndexExpr += "["
		case token.RBRACK:
			notIndexExpr += "]"
			return expr
		default:
			expr.Index = &injast.Ident{Name: lit, Location: paramLoc}
			notIndexExpr += lit
		}
	}
}
