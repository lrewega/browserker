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
		dec := json.NewDecoder(bytes.NewReader(b.body.Original))
		tok, err := dec.Token()
		if err == io.EOF {
			return b.body, nil
		}
		object := &injast.ObjectExpr{Location: browserk.InjectJSON, Fields: make([]browserk.InjectionExpr, 0), LPos: 0}
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
		pos := dec.InputOffset()
		if err == io.EOF {
			return
		}

		switch t := tok.(type) {
		case json.Delim:
			delim := rune(t)
			switch delim {
			case '{':
				subObject.EncChar = '{'
				subObject.LPos = browserk.InjectionPos(pos)
				b.parseJSONObject(dec, subObject, depth+1)
				object.Fields = append(object.Fields, subObject)
			case '}':
				kv.Value = subObject
				if depth == 0 {
					return
				}
			case '[':
				subObject.EncChar = '['
				subObject.LPos = browserk.InjectionPos(pos)
				b.parseJSONArray(dec, subObject, depth+1)
				kv.Value = subObject
				object.Fields = append(object.Fields, kv)
				kv = &injast.KeyValueExpr{}
			}
		case string:
			ident := &injast.Ident{
				NamePos: browserk.InjectionPos(int(pos) - 1 - len(t)),
				Name:    t,
				EncChar: '"',
			}

			if kv.Key == nil {
				ident.Location = browserk.InjectJSONName
				kv.Key = ident
				kv.Sep = browserk.InjectionPos(int(pos))
				kv.SepChar = ':'
			} else {
				ident.Location = browserk.InjectJSONValue
				kv.Value = ident
				object.Fields = append(object.Fields, kv)
				kv = &injast.KeyValueExpr{}
			}

		case float64:
			asStr := floatStr(t)
			ident := &injast.Ident{
				NamePos: browserk.InjectionPos(int(pos) - len(asStr)),
				Name:    strconv.FormatFloat(t, 'f', -1, 64),
			}
			// this probably shouldn't happen? (float as a name)
			if kv.Key == nil {
				ident.Location = browserk.InjectJSONName
				kv.Key = ident
				kv.Sep = browserk.InjectionPos(int(pos))
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
				NamePos: browserk.InjectionPos(int(pos) - len(asStr)),
				Name:    asStr,
			}
			// this probably shouldn't happen? (bool as a name)
			if kv.Key == nil {
				ident.Location = browserk.InjectJSONName
				kv.Key = ident
				kv.Sep = browserk.InjectionPos(int(pos))
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
				NamePos: browserk.InjectionPos(int(pos) - len(asStr)),
				Name:    asStr,
			}
			// this probably shouldn't happen? (null as a name)
			if kv.Key == nil {
				ident.Location = browserk.InjectJSONName
				kv.Key = ident
				kv.Sep = browserk.InjectionPos(int(pos))
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
		pos := dec.InputOffset()
		if err == io.EOF {
			return
		}

		switch t := tok.(type) {
		case json.Delim:

			delim := rune(t)
			switch delim {
			case '{':
				subObject.EncChar = '{'
				subObject.LPos = browserk.InjectionPos(pos)
				b.parseJSONObject(dec, subObject, depth+1)
			case '}':
			case '[':
				subObject.EncChar = '['
				subObject.LPos = browserk.InjectionPos(pos)
				b.parseJSONArray(dec, subObject, depth+1)
			case ']':
				return
			}
		case string:
			ident := &injast.Ident{
				NamePos:  browserk.InjectionPos(int(pos) - 2 - len(t)),
				Name:     t,
				EncChar:  '"',
				Location: browserk.InjectJSONValue,
			}
			object.Fields = append(object.Fields, ident)
		case float64:
			asStr := floatStr(t)
			ident := &injast.Ident{
				NamePos:  browserk.InjectionPos(int(pos) - len(asStr)),
				Name:     strconv.FormatFloat(t, 'f', -1, 64),
				Location: browserk.InjectJSONValue,
			}
			object.Fields = append(object.Fields, ident)
		case bool:
			asStr := "false"
			if t {
				asStr = "true"
			}
			ident := &injast.Ident{
				NamePos:  browserk.InjectionPos(int(pos) - len(asStr)),
				Name:     asStr,
				Location: browserk.InjectJSONValue,
			}
			object.Fields = append(object.Fields, ident)
		case nil:
			asStr := "null"
			ident := &injast.Ident{
				NamePos:  browserk.InjectionPos(int(pos) - len(asStr)),
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

func floatStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
