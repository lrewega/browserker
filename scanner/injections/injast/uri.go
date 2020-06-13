package injast

// URI for injecting into URI/query/fragments
type URI struct {
	Paths      []*Ident
	File       *Ident
	QueryDelim byte
	Query      *Query
	Fragment   *Fragment
	Original   []byte
	Modified   []byte
}

// NewURI for injection purposes
func NewURI(original []byte) *URI {
	return &URI{
		Original: original,
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

// Copy does a deep copy of the URI
func (u *URI) Copy() *URI {
	orig := append([]byte(nil), u.Original...)
	n := NewURI(orig)
	if u.Paths != nil && len(u.Paths) > 0 {
		n.Paths = make([]*Ident, len(u.Paths))
		for i, path := range u.Paths {
			n.Paths[i] = &Ident{NamePos: path.NamePos, Name: path.Name}
		}
	}

	if u.File != nil {
		n.File = &Ident{NamePos: u.File.NamePos, Name: u.File.Name}
	}

	n.QueryDelim = u.QueryDelim

	if u.Query.Params != nil && len(u.Query.Params) > 0 {
		n.Query.Params = make([]*KeyValueExpr, len(u.Query.Params))
		for i, param := range u.Query.Params {
			n.Query.Params[i] = CopyKeyValueExpr(param)
		}
	}

	if u.Fragment.Paths != nil && len(u.Fragment.Paths) > 0 {
		n.Fragment.Paths = make([]*Ident, len(u.Fragment.Paths))
		for i, path := range u.Fragment.Paths {
			n.Fragment.Paths[i] = &Ident{NamePos: path.NamePos, Name: path.Name}
		}
	}

	if u.Fragment.Params != nil && len(u.Fragment.Params) > 0 {
		n.Fragment.Params = make([]*KeyValueExpr, len(u.Fragment.Params))
		for i, param := range u.Fragment.Params {
			n.Fragment.Params[i] = CopyKeyValueExpr(param)
		}
	}

	return n
}

// String -ify the URI
func (u *URI) String() string {
	var lastPos Pos

	uri := "/"
	lastPos++

	if u.Paths != nil && len(u.Paths) > 0 {
		for _, p := range u.Paths {
			uri += p.String() + "/"
			lastPos = p.End() + 1 // add 1 for slash
		}
	}

	if u.File != nil && u.File.String() != "" {
		uri += u.File.String()
		lastPos = u.File.End()
	}

	if u.Query.Params != nil && len(u.Query.Params) > 0 {
		// add everything between  lastPos and firstParam
		firstParam := u.Query.Params[0].Pos()
		uri += string(u.Original[lastPos:firstParam])

		for i, p := range u.Query.Params {
			uri += p.String()
			if i+1 != len(u.Query.Params) {
				uri += "&"
			}
			lastPos = p.End()
		}
	}

	if u.Fragment.Paths != nil && len(u.Fragment.Paths) > 0 {
		// add everything between firstPath pos and lastPos
		firstPath := u.Fragment.Paths[0].Pos()
		uri += string(u.Original[lastPos:firstPath])

		for i, p := range u.Fragment.Paths {
			uri += p.String()
			if i+1 != len(u.Fragment.Paths) {
				uri += "/"
			}
			lastPos = p.End()
		}
	}

	if u.Fragment.Params != nil && len(u.Fragment.Params) > 0 {
		// add everything between firstPath pos and lastPos
		firstParam := u.Fragment.Params[0].Pos()
		uri += string(u.Original[lastPos:firstParam])
		for i, p := range u.Fragment.Params {
			if i != 0 {
				uri += string(u.Original[lastPos:p.Pos()])
			}
			uri += p.String()
			lastPos = p.End()
		}
	}

	// account for trailing delimiters
	if int(lastPos) != len(u.Original) {
		uri += string(u.Original[lastPos:len(u.Original)])
	}

	u.Modified = []byte(uri)
	return uri
}

// PathOnly part as a string
func (u *URI) PathOnly() string {
	uri := "/"
	if u.Paths != nil && len(u.Paths) > 0 {
		for _, p := range u.Paths {
			uri += p.String() + "/"
		}
	}
	return uri
}

// FileOnly as a string
func (u *URI) FileOnly() string {
	if u.File != nil {
		return u.File.String()
	}
	return ""
}

// Query part of a URI
type Query struct {
	Params []*KeyValueExpr
}

// Fragment part of a URI
type Fragment struct {
	Paths  []*Ident
	Params []*KeyValueExpr
}
