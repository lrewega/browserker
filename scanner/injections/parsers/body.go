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
	}

	return b.body, nil
}

func (b *BodyParser) parseApplicationURLEncoded(paramLoc browserk.InjectionLocation) {

	for {
		param := &injast.KeyValueExpr{SepChar: '='}
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
				if b.s.Peek() == '[' {
					param.Key = b.handleIndexExpr(pos, lit)
				} else {
					param.Key = &injast.Ident{NamePos: pos, Name: lit}
				}
				paramLoc = browserk.InjectBodyName
			} else {
				param.Value = &injast.Ident{NamePos: pos, Name: lit}
			}
			goto SCAN
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
