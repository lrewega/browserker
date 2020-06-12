package injast

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
}

type (

	// An Ident node represents an identifier.
	Ident struct {
		NamePos Pos    // identifier position
		Name    string // identifier name
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X      Expr // expression
		Lbrack Pos  // position of "["
		Index  Expr // index expression
		Rbrack Pos  // position of "]"
	}

	// A KeyValueExpr node represents (key : value) pairs
	// in composite literals.
	//
	KeyValueExpr struct {
		Key     Expr
		Sep     Pos  // position of separator
		SepChar rune // separator value
		Value   Expr
	}
)

func (*Ident) exprNode()  {}
func (x *Ident) Pos() Pos { return x.NamePos }
func (x *Ident) End() Pos { return Pos(int(x.NamePos) + len(x.Name)) }
func (id *Ident) String() string {
	if id != nil {
		return id.Name
	}
	return ""
}

func (*IndexExpr) exprNode()  {}
func (x *IndexExpr) Pos() Pos { return x.X.Pos() }
func (x *IndexExpr) End() Pos { return x.Rbrack + 1 }
func (x *IndexExpr) String() string {
	s := x.X.String()
	s += "["
	s += x.Index.String()
	s += "]"
	return s
}

func (*KeyValueExpr) exprNode()  {}
func (x *KeyValueExpr) Pos() Pos { return x.Key.Pos() }
func (x *KeyValueExpr) End() Pos { return x.Value.End() }
func (x *KeyValueExpr) String() string {
	s := x.Key.String()
	s += string(x.SepChar)
	s += x.Value.String()
	return s
}

func CopyExpr(e Expr) Expr {
	switch t := e.(type) {
	case *Ident:
		return &Ident{NamePos: t.NamePos, Name: t.Name}
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
		Key:     CopyExpr(kv.Key),
		Sep:     kv.Sep,
		SepChar: kv.SepChar,
		Value:   CopyExpr(kv.Value),
	}
}

func CopyIndexExpr(id *IndexExpr) *IndexExpr {
	return &IndexExpr{
		X:      CopyExpr(id.X),
		Lbrack: id.Lbrack,
		Index:  CopyExpr(id.Index),
		Rbrack: id.Rbrack,
	}
}
