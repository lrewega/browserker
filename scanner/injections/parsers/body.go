package parsers

import (
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/scanner"
	"gitlab.com/browserker/scanner/injections/token"
)

type BodyMode int

const (
	ApplicationURLEncoded BodyMode = iota + 1
	Multipart
	JSON
	XML
	GRAPHQL
)

type BodyParser struct {
	s       *scanner.Scanner
	mode    BodyMode
	kvMode  keyValueMode
	kvIndex int
	body    *injast.Body
}

// Parse body data
func (b *BodyParser) Parse(body []byte) (*injast.Body, error) {
	b.s = scanner.New()
	b.s.Init([]byte(body), scanner.Body)
	b.body = injast.NewBody(body)
	b.mode = ApplicationURLEncoded
	b.kvMode = keyMode
	b.kvIndex = 0
	if b.s.PeekIsBodyJSON() {
		b.mode = JSON
	} else if b.s.PeekIsBodyXML() {
		b.mode = XML
	}

	switch b.mode {
	case ApplicationURLEncoded:
		b.parseApplicationURLEncoded(browserk.InjectBody)
	case JSON:
		b.parseJSON(browserk.InjectJSON)
	}

	return b.body, nil
}

func (b *BodyParser) parseJSON(paramLoc browserk.InjectionLocation) {
	if b.s.PeekBackwards() == '[' {
		param := &injast.IndexExpr{Location: paramLoc}
		b.parseJSONArray(param, paramLoc)
		return
	}

	param := &injast.KeyValueExpr{SepChar: ':', Location: paramLoc}
	b.parseJSONKeyValue(param, paramLoc)
}

func (b *BodyParser) parseJSONArray(param *injast.IndexExpr, paramLoc browserk.InjectionLocation) {
	b.body.Fields = append(b.body.Fields, param)
	for {
		pos, tok, lit := b.s.Scan()

		switch tok {
		case token.EOF:
			return
		case token.RBRACK:
			param.Rbrack = pos
			return
		case token.LBRACE:
			kv := &injast.KeyValueExpr{Location: browserk.InjectJSONName}
			param.Index = kv
			b.parseJSONKeyValue(kv, browserk.InjectJSONName)
		case token.LBRACK:
			newArray := &injast.IndexExpr{Location: paramLoc, Lbrack: pos}
			param.Index = newArray
			b.parseJSONArray(newArray, paramLoc)
		default:
			if lit == "" {
				lit = tok.String()
			}
			if param.Index == nil {
				param.Index = &injast.Ident{NamePos: pos, Name: lit}
			} else if i, ok := param.Index.(*injast.Ident); ok {
				i.Name += lit
			}
		}
	}
}

func (b *BodyParser) parseJSONKeyValue(param *injast.KeyValueExpr, paramLoc browserk.InjectionLocation) {
	b.body.Fields = append(b.body.Fields, param)
	for {
		pos, tok, lit := b.s.Scan()

		switch tok {
		case token.EOF:
			return
		case token.LBRACE:
			if paramLoc == browserk.InjectJSONValue {
				kv := &injast.KeyValueExpr{Location: browserk.InjectJSONName}
				param.Value = kv
				b.parseJSONKeyValue(kv, browserk.InjectJSONName)
			}
		case token.LBRACK:
			if paramLoc == browserk.InjectJSONValue {
				newArray := &injast.IndexExpr{Location: paramLoc, Lbrack: pos}
				param.Value = newArray
				b.parseJSONArray(newArray, paramLoc)
			}
		case token.COLON:
			if paramLoc == browserk.InjectJSONName {
				paramLoc = browserk.InjectJSONValue
				param.Sep = pos
				param.SepChar = ':'
			}
		case token.COMMA:
			// oh shit
		default:
			if lit == "" {
				lit = tok.String()
			}
			if paramLoc == browserk.InjectJSONName {
				if param.Key == nil {
					param.Key = &injast.Ident{NamePos: pos, Name: lit}
				} else if k, ok := param.Key.(*injast.Ident); ok {
					k.Name += lit
				}
			} else if paramLoc == browserk.InjectJSONValue {
				if param.Value == nil {
					param.Value = &injast.Ident{NamePos: pos, Name: lit}
				} else if v, ok := param.Value.(*injast.Ident); ok {
					v.Name += lit
				}
			}
		}
	}
}

func (b *BodyParser) parseApplicationURLEncoded(paramLoc browserk.InjectionLocation) {

	for {
		param := &injast.KeyValueExpr{SepChar: '=', Location: paramLoc}
		b.body.Fields = append(b.body.Fields, param)
	SCAN:
		pos, tok, lit := b.s.Scan()
		if tok == token.EOF {
			return
		}
		switch tok {
		case token.AND:
			if paramLoc == browserk.InjectBodyValue {
				paramLoc = browserk.InjectBody
			}
		case token.ASSIGN:
			if paramLoc == browserk.InjectBodyName {
				paramLoc = browserk.InjectBodyValue
				param.Sep = pos
				goto SCAN
			}
		default:
			if paramLoc == browserk.InjectBody || paramLoc == browserk.InjectBodyName {
				if b.s.PeekBackwards() == '[' {
					param.Key = b.handleIndexExpr(pos, lit)
				} else {
					param.Key = &injast.Ident{NamePos: pos, Name: lit}
				}
				paramLoc = browserk.InjectBodyName
			} else {
				param.Value = b.handleValueExpr(pos, tok, lit)
			}
			goto SCAN
		}
	}
}

func (b *BodyParser) handleValueExpr(originalPos browserk.InjectionPos, originalTok token.Token, originalLit string) browserk.InjectionExpr {
	if originalLit == "" {
		originalLit = originalTok.String()
	}
	value := &injast.Ident{NamePos: originalPos, Name: originalLit}
	for {
		// short circuit so we don't consume the body name provided we don't start with a [ or ]
		if b.s.PeekBackwards() == '&' && (originalTok != token.LBRACK && originalTok != token.RBRACK) {
			return value
		}
		_, tok, lit := b.s.Scan()
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

func (b *BodyParser) handleIndexExpr(originalPos browserk.InjectionPos, lit string) browserk.InjectionExpr {
	paramLoc := browserk.InjectBodyIndex

	expr := &injast.IndexExpr{
		X:        &injast.Ident{NamePos: originalPos, Name: lit, Location: paramLoc},
		Lbrack:   0,
		Index:    nil,
		Rbrack:   0,
		Location: paramLoc,
	}

	notIndexExpr := lit

	paramLoc = browserk.InjectBodyName

	for {
		pos, tok, lit := b.s.Scan()
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
