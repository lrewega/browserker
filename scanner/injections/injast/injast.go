package injast

import (
	"bytes"
	"strings"

	"gitlab.com/browserker/browserk"
)

type (

	// An Ident node represents an identifier.
	Ident struct {
		NamePos  browserk.InjectionPos // identifier position
		Name     string                // identifier name
		Mod      string
		Modded   bool
		EncChar  rune // encapsulation char ' " { or [ etc
		Location browserk.InjectionLocation
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X        browserk.InjectionExpr // expression
		Lbrack   browserk.InjectionPos  // position of "["
		Index    browserk.InjectionExpr // index expression
		Rbrack   browserk.InjectionPos  // position of "]"
		Location browserk.InjectionLocation
	}

	// A KeyValueExpr node represents (key : value) pairs
	// in composite literals.
	//
	KeyValueExpr struct {
		Key      browserk.InjectionExpr
		Sep      browserk.InjectionPos // position of separator
		SepChar  rune                  // separator value
		Value    browserk.InjectionExpr
		Location browserk.InjectionLocation
	}

	// ObjectExpr represents an object (JSON/XML) with it's nested
	// fields Modifying it will replace *everything* with a Modified string
	// call Reset() to undo
	ObjectExpr struct {
		Fields   []browserk.InjectionExpr
		LPos     browserk.InjectionPos
		RPos     browserk.InjectionPos
		Location browserk.InjectionLocation
		Mod      string
		Modded   bool
		EncChar  rune // encapsulation character, { for objects, [ for arrays
	}
)

// Pos of this identifier
func (x *ObjectExpr) Pos() browserk.InjectionPos { return x.LPos }

// Loc of this object
func (x *ObjectExpr) Loc() browserk.InjectionLocation { return x.Location }

// End of this identifier
func (x *ObjectExpr) End() browserk.InjectionPos {
	return browserk.InjectionPos(int(x.LPos) + int(x.RPos))
}

// Reset any injection modifications
func (x *ObjectExpr) Reset() {
	x.Modded = false
	x.Mod = ""
}

func (x *ObjectExpr) String() string {
	if x == nil {
		return ""
	}

	if x.Modded {
		return x.Mod
	}
	if x.Fields == nil || len(x.Fields) == 0 {
		if x.EncChar == '{' {
			return "{}"
		} else if x.EncChar == '[' {
			return "[]"
		}
		return ""
	}

	encChar := '}'
	if x.EncChar == '[' {
		encChar = ']'
	}
	all := &bytes.Buffer{}
	all.WriteByte(byte(x.EncChar))
	asStrs := make([]string, len(x.Fields))
	for i, field := range x.Fields {
		asStrs[i] = field.String()
	}
	all.Write([]byte(strings.Join(asStrs, ", ")))
	all.WriteByte(byte(encChar))
	return string(all.Bytes())
}

// Modify sets a new field because End() and Pos() will be incorrect
// if we modify the Name field. All access should call String()
// so we can handle when a value is modified
func (x *ObjectExpr) Modify(newValue string) {
	x.Modded = true
	x.Mod = newValue
}

// Inject a nw value
func (x *ObjectExpr) Inject(newValue string, _ browserk.InjectionType) bool {
	x.Modify(newValue)
	return true
}

// Pos of this identifier
func (x *Ident) Pos() browserk.InjectionPos { return x.NamePos }

// End of this identifier
func (x *Ident) End() browserk.InjectionPos {
	return browserk.InjectionPos(int(x.NamePos) + len(x.Name))
}

// String the identfier, quoting it if necessary
func (x *Ident) String() string {
	if x == nil {
		return ""
	}
	quote := ""
	if x.EncChar != 0 {
		quote = string(x.EncChar)
	}
	if x.Modded {
		return quote + x.Mod + quote
	}
	return quote + x.Name + quote
}

// Modify sets a new field because End() and Pos() will be incorrect
// if we modify the Name field. All access should call String()
// so we can handle when a value is modified
func (x *Ident) Modify(newValue string) {
	x.Modded = true
	x.Mod = newValue
}

// Loc of this identifier
func (x *Ident) Loc() browserk.InjectionLocation { return x.Location }

// Inject a nw value
func (x *Ident) Inject(newValue string, _ browserk.InjectionType) bool {
	x.Modify(newValue)
	return true
}

// Reset any injection modifications
func (x *Ident) Reset() {
	x.Modded = false
	x.Mod = ""
}

// Pos position
func (x *IndexExpr) Pos() browserk.InjectionPos { return x.X.Pos() }

// End position
func (x *IndexExpr) End() browserk.InjectionPos { return x.Rbrack + 1 }

func (x *IndexExpr) String() string {
	s := x.X.String()
	s += "["
	if x.Index != nil {
		s += x.Index.String()
	}
	s += "]"
	return s
}

// Inject a new value of InjectionType (either index or value)
func (x *IndexExpr) Inject(newValue string, injType browserk.InjectionType) bool {
	if injType == browserk.InjectIndex {
		return x.Index.Inject(newValue, injType)
	}
	return x.X.Inject(newValue, injType)
}

// Reset any modifications
func (x *IndexExpr) Reset() {
	x.Index.Reset()
	x.X.Reset()
}

// Loc for injection
func (x *IndexExpr) Loc() browserk.InjectionLocation { return x.Location }

// Pos position
func (x *KeyValueExpr) Pos() browserk.InjectionPos { return x.Key.Pos() }

// End of entire KV pos
func (x *KeyValueExpr) End() browserk.InjectionPos { return x.Value.End() }

func (x *KeyValueExpr) String() string {
	s := ""
	if x.Key != nil {
		s = x.Key.String()
	}

	if x.SepChar != 0 {
		s += string(x.SepChar)
		// space after k: v
		if x.Location == browserk.InjectJSON {
			s += " "
		}
	}
	if x.Value != nil {
		s += x.Value.String()
	}
	return s
}

// Loc for injection
func (x *KeyValueExpr) Loc() browserk.InjectionLocation { return x.Location }

// Inject a new value of InjectionType
func (x *KeyValueExpr) Inject(newValue string, injType browserk.InjectionType) bool {
	if injType == browserk.InjectName {
		x.Key.Inject(newValue, injType)
	} else if injType == browserk.InjectValue {
		x.Value.Inject(newValue, injType)
	} else if injType == browserk.InjectIndex {
		if index, ok := x.Key.(*IndexExpr); ok {
			return index.Inject(newValue, injType)
		}
		return false
	}

	return true
}

// Reset any modifications
func (x *KeyValueExpr) Reset() {
	x.Key.Reset()
	x.Value.Reset()
}

// CopyExpr returns a deep copy
func CopyExpr(e browserk.InjectionExpr) browserk.InjectionExpr {
	switch t := e.(type) {
	case *Ident:
		return &Ident{NamePos: t.NamePos, Name: t.Name, Location: t.Location, EncChar: t.EncChar}
	case *IndexExpr:
		return CopyIndexExpr(t)
	case *KeyValueExpr:
		return CopyKeyValueExpr(t)
	case *ObjectExpr:
		return CopyObjectExpr(t)
	default:
		return nil
	}
}

// CopyKeyValueExpr returns a deep copy
func CopyKeyValueExpr(kv *KeyValueExpr) *KeyValueExpr {
	return &KeyValueExpr{
		Key:      CopyExpr(kv.Key),
		Sep:      kv.Sep,
		SepChar:  kv.SepChar,
		Value:    CopyExpr(kv.Value),
		Location: kv.Location,
	}
}

func CopyObjectExpr(o *ObjectExpr) *ObjectExpr {
	copiedFields := make([]browserk.InjectionExpr, len(o.Fields))
	for i := 0; i < len(o.Fields); i++ {
		copiedFields[i] = CopyExpr(o.Fields[i])
	}
	return &ObjectExpr{
		Fields:   o.Fields,
		LPos:     o.LPos,
		Location: o.Location,
	}
}

// CopyIndexExpr returns a deep copy
func CopyIndexExpr(id *IndexExpr) *IndexExpr {
	return &IndexExpr{
		X:        CopyExpr(id.X),
		Lbrack:   id.Lbrack,
		Index:    CopyExpr(id.Index),
		Rbrack:   id.Rbrack,
		Location: id.Location,
	}
}
