package injast

import (
	"gitlab.com/browserker/browserk"
)

// URIv2 for injecting into URIv2/query/fragments
type URIv2 struct {
	Paths      []*Ident
	File       *Ident
	QueryDelim byte
	Query      *Query
	Fragment   *Fragment
	Fields     []browserk.InjectionExpr
	Original   []byte
	Modified   []byte
}

// NewURIv2 for injection purposes
func NewURIv2(original []byte) *URIv2 {
	return &URIv2{
		Original: original,
		Fields:   make([]browserk.InjectionExpr, 0),
		Paths:    make([]*Ident, 0),
		File:     &Ident{},
		Query: &Query{
			Params: make([]*KeyValueExpr, 0),
		},
		Fragment: &Fragment{
			Paths:  make([]*Ident, 0),
			Params: make([]*KeyValueExpr, 0),
		},
	}
}

// Copy does a deep copy of the URIv2
func (u *URIv2) Copy() *URIv2 {
	orig := append([]byte(nil), u.Original...)
	n := NewURIv2(orig)

	if u.Fields != nil && len(u.Fields) > 0 {
		n.Fields = make([]browserk.InjectionExpr, len(u.Fields))
		for i, param := range u.Fields {
			n.Fields[i] = CopyExpr(param)
		}
	}

	return n
}

// String -ify the URIv2
func (u *URIv2) String() string {
	uri := ""
	for _, f := range u.Fields {
		uri += f.String()
	}
	u.Modified = []byte(uri)
	return uri
}

// PathOnly part as a string
func (u *URIv2) PathOnly() string {
	URIv2 := "/"
	for _, f := range u.Fields {
		if f.Loc() == browserk.InjectPath || f.Loc() == browserk.InjectFile {
			URIv2 += f.String()
		}
	}
	return URIv2
}

// FileOnly as a string
func (u *URIv2) FileOnly() string {
	for _, f := range u.Fields {
		if f.Loc() == browserk.InjectFile {
			return f.String()
		}
	}
	return ""
}

func (u *URIv2) HasParams() bool {
	return u.Query.Params != nil && len(u.Query.Params) > 0
}
