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

func NewBody(original []byte) *Body {
	return &Body{
		Original: original,
		Fields:   make([]browserk.InjectionExpr, 0),
	}
}
