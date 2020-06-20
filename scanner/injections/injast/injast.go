package injast

import "gitlab.com/browserker/browserk"

type Pos int

// All node types implement the Node interface.
type Node interface {
	Pos() Pos // position of first character belonging to the node
	End() Pos // position of first character immediately after the node
}

type Expr interface {
	Node
	exprNode()
	String() string
	Loc() browserk.InjectionLocation
}

type (

	// An Ident node represents an identifier.
	Ident struct {
		NamePos  Pos    // identifier position
		Name     string // identifier name
		Mod      string
		Modded   bool
		Location browserk.InjectionLocation
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X        Expr // expression
		Lbrack   Pos  // position of "["
		Index    Expr // index expression
		Rbrack   Pos  // position of "]"
		Location browserk.InjectionLocation
	}

	// A KeyValueExpr node represents (key : value) pairs
	// in composite literals.
	//
	KeyValueExpr struct {
		Key      Expr
		Sep      Pos  // position of separator
		SepChar  rune // separator value
		Value    Expr
		Location browserk.InjectionLocation
	}
)

func (*Ident) exprNode()  {}
func (x *Ident) Pos() Pos { return x.NamePos }
func (x *Ident) End() Pos { return Pos(int(x.NamePos) + len(x.Name)) }
func (x *Ident) String() string {
	if x != nil {
		if x.Modded {
			return x.Mod
		}
		return x.Name
	}
	return ""
}
func (x *Ident) Loc() browserk.InjectionLocation { return x.Location }

// Modify sets a new field because End() and Pos() will be incorrect
// if we modify the Name field. All access should call String()
// so we can handle when a value is modified
func (x *Ident) Modify(newValue string) {
	x.Modded = true
	x.Mod = newValue
}

func (*IndexExpr) exprNode()  {}
func (x *IndexExpr) Pos() Pos { return x.X.Pos() }
func (x *IndexExpr) End() Pos { return x.Rbrack + 1 }
func (x *IndexExpr) String() string {
	s := x.X.String()
	s += "["
	if x.Index != nil {
		s += x.Index.String()
	}
	s += "]"
	return s
}
func (x *IndexExpr) Loc() browserk.InjectionLocation { return x.Location }

func (*KeyValueExpr) exprNode()  {}
func (x *KeyValueExpr) Pos() Pos { return x.Key.Pos() }
func (x *KeyValueExpr) End() Pos { return x.Value.End() }
func (x *KeyValueExpr) String() string {
	s := x.Key.String()
	s += string(x.SepChar)
	s += x.Value.String()
	return s
}
func (x *KeyValueExpr) Loc() browserk.InjectionLocation { return x.Location }

func CopyExpr(e Expr) Expr {
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

func ReplaceExpr(e Expr, newVal, field string) bool {
	switch t := e.(type) {
	case *Ident:
		t.Modify(newVal)
		return true
	case *IndexExpr:
		if field == "key" {
			return ReplaceExpr(t.X, newVal, field)
		} else if field == "index" {
			return ReplaceExpr(t.Index, newVal, field)
		}
	case *KeyValueExpr:
		if field == "key" {
			return ReplaceExpr(t.Key, newVal, field)
		} else if field == "value" {
			return ReplaceExpr(t.Value, newVal, field)
		}
	default:
		return false
	}
	return false
}
