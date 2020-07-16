package injast

import (
	"gitlab.com/browserker/browserk"
)

// URI for injecting into URI/query/fragments
type URI struct {
	Fields   []browserk.InjectionExpr
	Original []byte
	Modified []byte
}

// NewURIv2 for injection purposes
func NewURIv2(original []byte) *URI {
	return &URI{
		Original: original,
		Fields:   make([]browserk.InjectionExpr, 0),
	}
}

// Copy does a deep copy of the URIv2
func (u *URI) Copy() *URI {
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
func (u *URI) String() string {
	uri := "/"
	setQuery := false
	setFrag := false

	for _, f := range u.Fields {
		switch f.Loc() {
		case browserk.InjectPath:
			uri += f.String() + "/"
			continue
		case browserk.InjectQuery, browserk.InjectQueryName:
			if setQuery == false {
				uri += "?"
				setQuery = true
			} else {
				uri += "&"
			}
		case browserk.InjectFragment, browserk.InjectFragmentName:
			if setFrag == false {
				uri += "#"
				setFrag = true
			}
		}
		uri += f.String()
	}
	u.Modified = []byte(uri)
	return uri
}

// PathOnly part as a string
func (u *URI) PathOnly() string {
	URIv2 := ""
	for _, f := range u.Fields {
		if f.Loc() == browserk.InjectPath {
			URIv2 += "/" + f.String()
		}
	}

	URIv2 += "/"
	return URIv2
}

// FileOnly as a string
func (u *URI) FileOnly() string {
	for _, f := range u.Fields {
		if f.Loc() == browserk.InjectFile {
			return f.String()
		}
	}
	return ""
}

func (u *URI) HasParams() bool {
	for _, f := range u.Fields {
		if f.Loc() == browserk.InjectQuery {
			return true
		}
	}
	return false
}
