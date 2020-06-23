package injast

import "gitlab.com/browserker/browserk"

type (

	// An Ident node represents an identifier.
	Ident struct {
		NamePos  browserk.InjectionPos // identifier position
		Name     string                // identifier name
		Mod      string
		Modded   bool
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
)

func (x *Ident) Pos() browserk.InjectionPos { return x.NamePos }
func (x *Ident) End() browserk.InjectionPos {
	return browserk.InjectionPos(int(x.NamePos) + len(x.Name))
}
func (x *Ident) String() string {
	if x != nil {
		if x.Modded {
			return x.Mod
		}
		return x.Name
	}
	return ""
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
func (x *Ident) Inject(newValue string, injType browserk.InjectionType) bool {
	x.Modify(newValue)
	return true
}

func (x *Ident) Reset() {
	x.Modded = false
	x.Mod = ""
}

func (x *IndexExpr) Pos() browserk.InjectionPos { return x.X.Pos() }
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
func (x *IndexExpr) Inject(newValue string, injType browserk.InjectionType) bool {
	if injType == browserk.InjectIndex {
		return x.Index.Inject(newValue, injType)
	}
	return x.X.Inject(newValue, injType)
}

func (x *IndexExpr) Reset() {
	x.Index.Reset()
	x.X.Reset()
}

func (x *IndexExpr) Loc() browserk.InjectionLocation { return x.Location }

func (x *KeyValueExpr) Pos() browserk.InjectionPos { return x.Key.Pos() }
func (x *KeyValueExpr) End() browserk.InjectionPos { return x.Value.End() }
func (x *KeyValueExpr) String() string {
	s := x.Key.String()
	s += string(x.SepChar)
	s += x.Value.String()
	return s
}
func (x *KeyValueExpr) Loc() browserk.InjectionLocation { return x.Location }

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

func (x *KeyValueExpr) Reset() {
	x.Key.Reset()
	x.Value.Reset()
}

func CopyExpr(e browserk.InjectionExpr) browserk.InjectionExpr {
	switch t := e.(type) {
	case *Ident:
		return &Ident{NamePos: t.NamePos, Name: t.Name, Location: t.Location}
	case *IndexExpr:
		return CopyIndexExpr(t)
	case *KeyValueExpr:
		return CopyKeyValueExpr(t)
	default:
		return nil
	}
}

func CopyKeyValueExpr(kv *KeyValueExpr) *KeyValueExpr {
	return &KeyValueExpr{
		Key:      CopyExpr(kv.Key),
		Sep:      kv.Sep,
		SepChar:  kv.SepChar,
		Value:    CopyExpr(kv.Value),
		Location: kv.Location,
	}
}

func CopyIndexExpr(id *IndexExpr) *IndexExpr {
	return &IndexExpr{
		X:        CopyExpr(id.X),
		Lbrack:   id.Lbrack,
		Index:    CopyExpr(id.Index),
		Rbrack:   id.Rbrack,
		Location: id.Location,
	}
}
