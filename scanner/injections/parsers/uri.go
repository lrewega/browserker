package parsers

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
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

// Parse a uri into it's parts
// TODO: Clean this up it's a disaster
func (u *URIParser) Parse(uri string) (*injast.URI, error) {
	log.Debug().Str("uri", uri).Msg("parsing for injection")
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
				// for urls that immediately end
				// eg: /path/file
				peek := u.s.Peek()
				// for urls that have query params, since we'll skip over the token
				// eg: /path/file?x=1
				peekBackwards := u.s.PeekBackwards()

				// handle case of /file/?blah first
				if peek == '?' && peekBackwards == '/' {
					p := &injast.Ident{
						NamePos:  pos,
						Name:     lit,
						Location: browserk.InjectPath,
					}
					u.uri.Paths = append(u.uri.Paths, p)
					u.uri.Fields = append(u.uri.Fields, p)
					// if the next char is a ?, ;, # or EOF that means this ident is a file part
				} else if (peek == '?' || peek == '#' || peek == ';' || peek == 0) ||
					(peekBackwards == '?' || peekBackwards == '#' || peekBackwards == ';' || peekBackwards == 0) {

					f := &injast.Ident{
						NamePos:  pos,
						Name:     lit,
						Location: browserk.InjectFile,
					}
					u.uri.File = f
					u.uri.Fields = append(u.uri.Fields, f)
				} else {
					p := &injast.Ident{
						NamePos:  pos,
						Name:     lit,
						Location: browserk.InjectPath,
					}
					u.uri.Paths = append(u.uri.Paths, p)
					u.uri.Fields = append(u.uri.Fields, p)
				}
			}
		// case File: file is a one shot, added under case Path so no need to capture here
		case Query:
			if tok == token.QUESTION {
				u.uri.QueryDelim = '?'
			}
			u.handleParams(tok, pos, lit, &u.uri.Query.Params, browserk.InjectQuery)
		case Fragment:
			switch tok {
			case token.HASH:
				continue
			case token.SLASH:
				u.mode = FragmentPath
				continue
			case token.IDENT:
				peek := u.s.Peek()
				back := u.s.PeekBackwards()
				if (peek == '?' || peek == '&' || peek == 0) || back == '/' {
					p := &injast.Ident{
						NamePos:  pos,
						Name:     lit,
						Location: browserk.InjectFragmentPath,
					}
					u.uri.Fragment.Paths = append(u.uri.Fragment.Paths, p)
					u.uri.Fields = append(u.uri.Fields, p)
					continue
				}
				u.kvIndex = 0 // reset kvIndex for fragment params
				u.kvMode = keyMode
				u.mode = FragmentQuery
				u.handleParams(tok, pos, lit, &u.uri.Fragment.Params, browserk.InjectFragment)
			}
		case FragmentPath:
			if tok == token.SLASH {
				continue
			} else if tok.IsLiteral() {
				p := &injast.Ident{
					NamePos:  pos,
					Name:     lit,
					Location: browserk.InjectFragmentPath,
				}
				u.uri.Fragment.Paths = append(u.uri.Fragment.Paths, p)
				u.uri.Fields = append(u.uri.Fields, p)

			}
		case FragmentQuery:
			u.handleParams(tok, pos, lit, &u.uri.Fragment.Params, browserk.InjectFragment)
		}
	}
}

func (u *URIParser) handleParams(tok token.Token, pos browserk.InjectionPos, lit string, params *[]*injast.KeyValueExpr, loc browserk.InjectionLocation) {
	paramLoc := browserk.InjectQueryName
	if loc == browserk.InjectQuery && u.kvMode == keyMode {
		paramLoc = browserk.InjectQueryName
	} else if loc == browserk.InjectQuery && u.kvMode == valueMode {
		paramLoc = browserk.InjectQueryValue
	} else if loc == browserk.InjectFragment && u.kvMode == keyMode {
		paramLoc = browserk.InjectFragmentName
	} else if loc == browserk.InjectFragment && u.kvMode == valueMode {
		paramLoc = browserk.InjectFragmentValue
	}

	switch tok {
	case token.ASSIGN:
		// /file?=asdf (invalid, but we must account for it)
		if len(*params) == 0 {
			kv := &injast.KeyValueExpr{
				Key:      &injast.Ident{NamePos: pos, Name: lit, Location: paramLoc},
				Sep:      0,
				SepChar:  0,
				Value:    nil,
				Location: loc,
			}
			*params = append(*params, kv)
			u.uri.Fields = append(u.uri.Fields, kv)
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

	if lit == "" {
		lit = tok.String()
	}

	if u.kvMode == keyMode {
		var key browserk.InjectionExpr
		key = &injast.Ident{NamePos: pos, Name: lit, Location: paramLoc}
		peek := u.s.PeekBackwards()
		if peek == '[' {
			key = u.handleIndexExpr(pos, lit, loc)
		}
		kv := &injast.KeyValueExpr{
			Key:      key,
			Sep:      pos,
			SepChar:  0,
			Value:    nil,
			Location: loc,
		}
		*params = append(*params, kv)
		u.uri.Fields = append(u.uri.Fields, kv)
		// make sure we don't have a nil value hanging off the KV
		if u.s.Peek() == 0 {
			kv.Value = &injast.Ident{NamePos: key.End()}
		}
	} else {
		(*params)[u.kvIndex].Value = u.handleValueExpr(pos, tok, lit)
	}
}

func (u *URIParser) handleValueExpr(originalPos browserk.InjectionPos, originalTok token.Token, originalLit string) browserk.InjectionExpr {
	if originalLit == "" {
		originalLit = originalTok.String()
	}
	value := &injast.Ident{NamePos: originalPos, Name: originalLit}
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

func (u *URIParser) handleIndexExpr(originalPos browserk.InjectionPos, lit string, loc browserk.InjectionLocation) browserk.InjectionExpr {
	paramLoc := browserk.InjectQueryIndex

	if loc == browserk.InjectFragment {
		paramLoc = browserk.InjectFragmentIndex
	}

	expr := &injast.IndexExpr{
		X:        &injast.Ident{NamePos: originalPos, Name: lit, Location: paramLoc},
		Lbrack:   0,
		Index:    nil,
		Rbrack:   0,
		Location: paramLoc,
	}

	notIndexExpr := lit

	paramLoc = browserk.InjectQueryName
	if loc == browserk.InjectFragment {
		paramLoc = browserk.InjectFragmentName
	}

	for {
		pos, tok, lit := u.s.Scan()
		switch tok {
		case token.EOF, token.ASSIGN:
			return &injast.Ident{NamePos: originalPos, Name: notIndexExpr, Location: paramLoc}
		case token.LBRACK:
			notIndexExpr += "["
			expr.Lbrack = pos
		case token.RBRACK:
			notIndexExpr += "]"
			expr.Rbrack = pos
			return expr
		default:
			expr.Index = &injast.Ident{NamePos: pos, Name: lit, Location: paramLoc}
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
		u.kvIndex = 0      // reset index
		u.kvMode = keyMode // reset kv mode to key
		u.mode = FragmentQuery
	}
}
