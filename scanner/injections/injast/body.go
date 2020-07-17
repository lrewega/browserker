package injast

import (
	"gitlab.com/browserker/browserk"
)

// Body for injecting into HTTP Body
type Body struct {
	Fields   []browserk.InjectionExpr
	Original []byte
	Modified []byte
}

// NewBody for injections into requests that have a request body
func NewBody(original []byte) *Body {
	return &Body{
		Original: original,
		Fields:   make([]browserk.InjectionExpr, 0),
	}
}

func (u *Body) Copy() *Body {
	orig := append([]byte(nil), u.Original...)
	n := NewBody(orig)

	if u.Fields != nil && len(u.Fields) > 0 {
		n.Fields = make([]browserk.InjectionExpr, len(u.Fields))
		for i, param := range u.Fields {
			n.Fields[i] = CopyExpr(param)
		}
	}

	return n
}

// IsJSON ?
func (b *Body) IsJSON() bool {
	return (b.Fields != nil && len(b.Fields) > 0 && b.Fields[0].Loc() == browserk.InjectJSON)
}

func (b *Body) String() string {
	res := ""
	if b == nil || b.Fields == nil || len(b.Fields) == 0 {
		return res
	}
	for i, f := range b.Fields {
		res += f.String()

		if f.Loc() == browserk.InjectBody && i+1 != len(b.Fields) {
			res += "&"
		}
	}
	return res
}
