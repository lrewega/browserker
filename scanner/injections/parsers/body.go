package parsers

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		pos := dec.InputOffset()
		if err == io.EOF {
			return b.body, nil
		}
		object := &injast.ObjectExpr{Location: browserk.InjectJSON, Fields: make([]browserk.InjectionExpr, 0), LPos: pos}
		if tok == json.Delim('[') {
			object.EncChar = '['
			b.parseJSONArray(dec, object, browserk.InjectJSON)
		} else if tok == json.Delim('{') {
			object.EncChar = '{'
			b.parseJSONObject(dec, object, browserk.InjectJSON)
		}

	}

	return b.body, nil
}

func (b *BodyParser) parseJSONObject(dec *json.Decoder, object *injast.ObjectExpr, paramLoc browserk.InjectionLocation) {
	for {
		tok, err := dec.Token()
		pos := dec.InputOffset()
		if err == io.EOF {
			return
		}

		switch t := tok.(type) {
		case json.Delim:
			if tok == json.Delim('{') {

				fmt.Printf("in object @ %d\n", pos)
			} else if tok == json.Delim('[') {
				fmt.Printf("in array @ %d\n", pos)
			}
		case string:
			// sub 2 for ": and sub length of str
			fmt.Printf("in string %v @ %d-%d\n", t, int(pos)-2-len(t), pos)
		case float64:
			fmt.Printf("in float64 %v @ %d-%d\n", t, int(pos)-floatLen(t), pos)
		case bool:
			fmt.Printf("in bool %v @ %d-%d\n", t, pos-4, pos)
		case nil:
			fmt.Printf("in null %v @ %d-%d\n", t, pos-4, pos) // pos-len(null)
		}
	}
}

func (b *BodyParser) parseJSONArray(dec *json.Decoder, object *injast.ObjectExpr, paramLoc browserk.InjectionLocation) {
	for {
		tok, err := dec.Token()
		pos := dec.InputOffset()
		if err == io.EOF {
			return
		}

		switch t := tok.(type) {
		case json.Delim:
			if tok == json.Delim('{') {

				fmt.Printf("in object @ %d\n", pos)
			} else if tok == json.Delim('[') {
				fmt.Printf("in array @ %d\n", pos)
			}
		case string:
			// sub 2 for ": and sub length of str
			fmt.Printf("in string %v @ %d-%d\n", t, int(pos)-2-len(t), pos)
		case float64:
			fmt.Printf("in float64 %v @ %d-%d\n", t, int(pos)-floatLen(t), pos)
		case bool:
			fmt.Printf("in bool %v @ %d-%d\n", t, pos-4, pos)
		case nil:
			fmt.Printf("in null %v @ %d-%d\n", t, pos-4, pos) // pos-len(null)
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

func floatLen(f float64) int {
	c := strconv.FormatFloat(f, 'f', -1, 64)
	fmt.Printf("formatted; %s\n", c)
	return len(c)
}

/*
func (b *BodyParser) parseJSONObject(param *injast.ObjectExpr, paramLoc browserk.InjectionLocation) {
	b.body.Fields = append(b.body.Fields, param)
	for {
		pos, tok, _ := b.s.Scan()

		switch tok {
		case token.RBRACE, token.RBRACK:
			param.RPos = pos
			return
		case token.EOF:
			return
		case token.LBRACK:
			if paramLoc == browserk.InjectJSON {
				paramLoc = browserk.InjectJSONName
			}
			arr := &injast.ObjectExpr{Location: paramLoc, LPos: pos, EncChar: '['}
			param.Fields = append(param.Fields, arr)
			b.parseJSONArray(arr, paramLoc)
		case token.LBRACE:
			if paramLoc == browserk.InjectJSON {
				paramLoc = browserk.InjectJSONName
			}
			kv := &injast.KeyValueExpr{Location: paramLoc}
			param.EncChar = '{'
			param.LPos = pos
			param.Fields = append(param.Fields, kv)
			key := &injast.Ident{}
			namePos := b.parseJSONName(key) // should return after : seperator
			kv.Key = key
			kv.Sep = namePos
			kv.SepChar = ':'
			// we don't know what type of object it is yet
			kv.Value, param.RPos = b.parseJSONValue()

		}
	}
}

// we don't know what type of value this is yet, could be KeyValue, Object, or Ident
func (b *BodyParser) parseJSONValue() (browserk.InjectionExpr, browserk.InjectionPos) {
	var v browserk.InjectionExpr

	for {
		pos, tok, lit := b.s.Scan()

		switch tok {
		case token.COLON:
			if v == nil {
				continue
			}
			ident, _ := v.(*injast.Ident)
			ident.Name += ":"
		case token.EOF:
			return v, pos
		case token.COMMA:
			return v, pos
		case token.LBRACE:
			o := &injast.ObjectExpr{Location: browserk.InjectJSONValue, LPos: pos, EncChar: '{'}
			kv := &injast.KeyValueExpr{Location: browserk.InjectJSON}
			o.Fields = append(o.Fields, kv)
			key := &injast.Ident{}
			b.parseJSONName(key)
			kv.Key = key
			kv.Value, o.RPos = b.parseJSONValue()
			v = o
		case token.LBRACK:
			o := &injast.ObjectExpr{Location: browserk.InjectJSONValue, LPos: pos, EncChar: '['}
			v = o
		case token.SQUOTE:
			if v == nil {
				v = &injast.Ident{
					Name:      lit,
					QuoteChar: '\'',
					QuotePos:  pos,
					Location:  browserk.InjectJSONValue,
				}
				continue
			}
			ident, _ := v.(*injast.Ident)
			ident.Name += "'"
		case token.DQUOTE:
			if v == nil {
				v = &injast.Ident{
					Name:      lit,
					QuoteChar: '"',
					QuotePos:  pos,
					Location:  browserk.InjectJSONValue,
				}
				continue
			}
			ident, _ := v.(*injast.Ident)
			ident.Name += "\""
		case token.SPACE:
			if v == nil {
				continue
			}
			ident, _ := v.(*injast.Ident)
			ident.Name += " "
		default:
			if v == nil {
				v = &injast.Ident{
					NamePos:  pos,
					Name:     lit,
					Location: browserk.InjectJSONValue,
				}
			}
			ident, _ := v.(*injast.Ident)
			if lit == "" {
				lit = tok.String()
			}
			ident.Name += lit
		}
	}
}

func (b *BodyParser) parseJSONArray(param *injast.ObjectExpr, paramLoc browserk.InjectionLocation) {
	b.body.Fields = append(b.body.Fields, param)
	for {
		pos, tok, _ := b.s.Scan()

		switch tok {
		case token.EOF:
			return
		case token.RBRACK:
			param.RPos = pos
			return
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
			if paramLoc == browserk.InjectJSON {
				paramLoc = browserk.InjectJSONName
			} else if paramLoc == browserk.InjectJSONValue {
				kv := &injast.KeyValueExpr{Location: browserk.InjectJSONName}
				param.Value = kv
				b.parseJSONKeyValue(kv, browserk.InjectJSONName)
			}
		case token.LBRACK:
			if paramLoc == browserk.InjectJSONValue {
				newArray := &injast.ObjectExpr{
					Fields:   make([]browserk.InjectionExpr, 0),
					LPos:     pos,
					Location: paramLoc,
					Mod:      "",
					Modded:   false,
					EncChar:  '[',
				}
				param.Value = newArray
				b.parseJSONArray(newArray, paramLoc)
			}
		case token.COLON:
			if paramLoc == browserk.InjectJSONName {
				paramLoc = browserk.InjectJSONValue
				param.Sep = pos
				param.SepChar = ':'
			}
		case token.DQUOTE, token.SQUOTE:
			if paramLoc == browserk.InjectJSONName || paramLoc == browserk.InjectJSON {
				if param.Key == nil {
					paramLoc = browserk.InjectJSONName
					key := &injast.Ident{QuotePos: pos, QuoteChar: rune(tok.String()[0]), Location: browserk.InjectJSONName}
					b.parseJSONName(key)
					param.Key = key
				}
			} else if paramLoc == browserk.InjectJSONValue {
				if param.Value == nil {
					param.Value = &injast.Ident{QuotePos: pos, QuoteChar: rune(tok.String()[0])}
				}
			}
		case token.COMMA:
			nextParam := &injast.KeyValueExpr{Location: browserk.InjectJSONName}
			b.parseJSONKeyValue(nextParam, browserk.InjectJSONName)
		default:
			if lit == "" {
				lit = tok.String()
			}
			if paramLoc == browserk.InjectJSONName || paramLoc == browserk.InjectJSON {
				// we don't capture quotes encapsulating the name
				if param.Key == nil {
					param.Key = &injast.Ident{NamePos: pos, Name: lit}
				} else if k, ok := param.Key.(*injast.Ident); ok {
					k.Name += lit
				}
			} else if paramLoc == browserk.InjectJSONValue {
				// we don't capture quotes encapsulating the value
				if param.Value == nil {
					param.Value = &injast.Ident{NamePos: pos, Name: lit}
				} else if v, ok := param.Value.(*injast.Ident); ok {
					v.Name += lit
				}
			}
		}
	}
}

func (b *BodyParser) parseJSONName(ident *injast.Ident) browserk.InjectionPos {
	for {
		pos, tok, lit := b.s.Scan()
		peek := b.s.PeekBackwards()
		switch tok {
		case token.EOF:
			return pos
		case token.DQUOTE:

			if ident.QuoteChar == '"' && peek == '0' || peek == ':' {
				return pos
			} else if ident.Name == "" && ident.QuoteChar == 0 {
				ident.QuoteChar = '"'
				ident.QuotePos = pos
			} else if ident.Name != "" {
				ident.Name += "\""
			}
		case token.SQUOTE:
			if ident.QuoteChar == '\'' && peek == '0' || peek == ':' {
				return pos
			} else if ident.Name == "" && ident.QuoteChar == 0 {
				ident.QuoteChar = '\''
				ident.QuotePos = pos
			} else if ident.Name != "" {
				ident.Name += "'"
			}
		default:
			if ident.NamePos == 0 {
				ident.NamePos = pos
			}
			ident.Name += lit
		}
	}
}

*/
