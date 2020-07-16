package parsers

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"

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
	s    *scanner.Scanner
	mode BodyMode
	body *injast.Body
}

// Parse body data
func (b *BodyParser) Parse(body []byte) (*injast.Body, error) {
	b.s = scanner.New()
	b.s.Init([]byte(body), scanner.Body)
	b.body = injast.NewBody(body)
	b.mode = ApplicationURLEncoded
	if b.s.PeekIsBodyJSON() {
		b.mode = JSON
	} else if b.s.PeekIsBodyXML() {
		b.mode = XML
	}

	switch b.mode {
	case ApplicationURLEncoded:
		b.parseApplicationURLEncoded(browserk.InjectBody)
	case JSON:
		dec := json.NewDecoder(bytes.NewReader(b.body.Original))
		tok, err := dec.Token()
		if err == io.EOF {
			return b.body, nil
		}
		object := &injast.ObjectExpr{Location: browserk.InjectJSON, Fields: make([]browserk.InjectionExpr, 0)}
		if tok == json.Delim('[') {
			object.EncChar = '['
			b.parseJSONArray(dec, object, 0)
		} else if tok == json.Delim('{') {
			object.EncChar = '{'
			b.parseJSONObject(dec, object, 0)
		}
		b.body.Fields = append(b.body.Fields, object)
	}

	return b.body, nil
}

func (b *BodyParser) parseJSONObject(dec *json.Decoder, object *injast.ObjectExpr, depth int) {
	kv := &injast.KeyValueExpr{}
	subObject := &injast.ObjectExpr{Location: browserk.InjectJSON, Fields: make([]browserk.InjectionExpr, 0)}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return
		}

		switch t := tok.(type) {
		case json.Delim:
			delim := rune(t)
			switch delim {
			case '{':
				subObject.EncChar = '{'
				b.parseJSONObject(dec, subObject, depth+1)
				object.Fields = append(object.Fields, subObject)
			case '}':
				kv.Value = subObject
				if depth == 0 {
					return
				}
			case '[':
				subObject.EncChar = '['
				b.parseJSONArray(dec, subObject, depth+1)
				kv.Value = subObject
				object.Fields = append(object.Fields, kv)
				kv = &injast.KeyValueExpr{}
			}
		case string:
			ident := &injast.Ident{
				Name:    t,
				EncChar: '"',
			}

			if kv.Key == nil {
				ident.Location = browserk.InjectJSONName
				kv.Key = ident
				kv.SepChar = ':'
			} else {
				ident.Location = browserk.InjectJSONValue
				kv.Value = ident
				object.Fields = append(object.Fields, kv)
				kv = &injast.KeyValueExpr{}
			}

		case float64:
			ident := &injast.Ident{
				Name: floatStr(t),
			}
			// this probably shouldn't happen? (float as a name)
			if kv.Key == nil {
				ident.Location = browserk.InjectJSONName
				kv.Key = ident
				kv.SepChar = ':'
			} else {
				ident.Location = browserk.InjectJSONValue
				kv.Value = ident
				object.Fields = append(object.Fields, kv)
				kv = &injast.KeyValueExpr{}
			}
		case bool:
			asStr := "false"
			if t {
				asStr = "true"
			}
			ident := &injast.Ident{
				Name: asStr,
			}
			// this probably shouldn't happen? (bool as a name)
			if kv.Key == nil {
				ident.Location = browserk.InjectJSONName
				kv.Key = ident
				kv.SepChar = ':'
			} else {
				ident.Location = browserk.InjectJSONValue
				kv.Value = ident
				object.Fields = append(object.Fields, kv)
				kv = &injast.KeyValueExpr{}
			}
		case nil:
			asStr := "null"
			ident := &injast.Ident{
				Name: asStr,
			}
			// this probably shouldn't happen? (null as a name)
			if kv.Key == nil {
				ident.Location = browserk.InjectJSONName
				kv.Key = ident
				kv.SepChar = ':'
			} else {
				ident.Location = browserk.InjectJSONValue
				kv.Value = ident
				object.Fields = append(object.Fields, kv)
				kv = &injast.KeyValueExpr{}
			}
		}
	}
}

func (b *BodyParser) parseJSONArray(dec *json.Decoder, object *injast.ObjectExpr, depth int) {
	subObject := &injast.ObjectExpr{Location: browserk.InjectJSON, Fields: make([]browserk.InjectionExpr, 0)}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return
		}

		switch t := tok.(type) {
		case json.Delim:

			delim := rune(t)
			switch delim {
			case '{':
				subObject.EncChar = '{'
				b.parseJSONObject(dec, subObject, depth+1)
			case '}':
			case '[':
				subObject.EncChar = '['
				b.parseJSONArray(dec, subObject, depth+1)
			case ']':
				return
			}
		case string:
			ident := &injast.Ident{
				Name:     t,
				EncChar:  '"',
				Location: browserk.InjectJSONValue,
			}
			object.Fields = append(object.Fields, ident)
		case float64:
			ident := &injast.Ident{
				Name:     floatStr(t),
				Location: browserk.InjectJSONValue,
			}
			object.Fields = append(object.Fields, ident)
		case bool:
			asStr := "false"
			if t {
				asStr = "true"
			}
			ident := &injast.Ident{
				Name:     asStr,
				Location: browserk.InjectJSONValue,
			}
			object.Fields = append(object.Fields, ident)
		case nil:
			asStr := "null"
			ident := &injast.Ident{
				Name:     asStr,
				Location: browserk.InjectJSONValue,
			}
			object.Fields = append(object.Fields, ident)
		}
	}
}

func (b *BodyParser) parseApplicationURLEncoded(paramLoc browserk.InjectionLocation) {

	for {
		param := &injast.KeyValueExpr{SepChar: '=', Location: paramLoc}
		b.body.Fields = append(b.body.Fields, param)
	SCAN:
		_, tok, lit := b.s.Scan()
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
				goto SCAN
			}
		default:
			if paramLoc == browserk.InjectBody || paramLoc == browserk.InjectBodyName {
				if b.s.PeekBackwards() == '[' {
					param.Key = b.handleIndexExpr(lit)
				} else {
					param.Key = &injast.Ident{Name: lit}
				}
				paramLoc = browserk.InjectBodyName
			} else {
				param.Value = b.handleValueExpr(tok, lit)
			}
			goto SCAN
		}
	}
}

func (b *BodyParser) handleValueExpr(originalTok token.Token, originalLit string) browserk.InjectionExpr {
	if originalLit == "" {
		originalLit = originalTok.String()
	}
	value := &injast.Ident{Name: originalLit}
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

func (b *BodyParser) handleIndexExpr(lit string) browserk.InjectionExpr {
	paramLoc := browserk.InjectBodyIndex

	expr := &injast.IndexExpr{
		X:        &injast.Ident{Name: lit, Location: paramLoc},
		Index:    nil,
		Location: paramLoc,
	}

	notIndexExpr := lit

	paramLoc = browserk.InjectBodyName

	for {
		_, tok, lit := b.s.Scan()
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

func floatStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
