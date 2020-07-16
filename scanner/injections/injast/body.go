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

// IsJSON ?
func (b *Body) IsJSON() bool {
	return (b.Fields != nil && len(b.Fields) > 0 && b.Fields[0].Loc() == browserk.InjectJSON)
}

func (b *Body) String() string {
	res := ""
	for _, f := range b.Fields {
		res += f.String()
	}
	return res
}
