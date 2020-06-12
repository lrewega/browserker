package injast

// URI for injecting into URI/query/fragments
type URI struct {
	Paths    []*Ident
	File     *Ident
	Query    *Query
	Fragment *Fragment
	Original []byte
	Modified []byte
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
			File:   &Ident{},
			Paths:  make([]*Ident, 0),
			Params: make([]*KeyValueExpr, 0),
		},
	}
}

// Copy does a deep copy of the URI
func (u *URI) Copy() *URI {
	n := NewURI(u.Original)
	if u.Paths != nil && len(u.Paths) > 0 {
		n.Paths = make([]*Ident, len(u.Paths))
		for i, path := range u.Paths {
			n.Paths[i] = &Ident{NamePos: path.NamePos, Name: path.Name}
		}
	}

	if u.File != nil {
		n.File = &Ident{NamePos: u.File.NamePos, Name: u.File.Name}
	}

	if u.Query.Params != nil && len(u.Query.Params) > 0 {
		n.Query.Params = make([]*KeyValueExpr, len(u.Query.Params))
		for i, param := range u.Query.Params {
			n.Query.Params[i] = CopyKeyValueExpr(param)
		}
	}

	if u.Fragment.File != nil {
		u.Fragment.File = &Ident{NamePos: u.Fragment.File.NamePos, Name: u.Fragment.File.Name}
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
	// we'll want to use the positional information to inject our changes, and then
	// re-encode
	/*
		uri := u.PathOnly()
		uri += u.FileOnly()

		if u.Query.Params != nil && len(u.Query.Params) > 0 {
			uri += "?"
			for i, p := range u.Query.Params {
				uri += p.String()
				if i+1 != len(u.Query.Params) {
					uri += "&"
				}
			}
		}
		if u.Fragment.Paths != nil && len(u.Fragment.Paths) > 0 {
			uri += "#"
			for i, p := range u.Fragment.Paths {
				uri += p.String()
				if i+1 != len(u.Fragment.Paths) {
					uri += "/"
				}
			}
		}
		if u.Fragment.File != nil {
			uri += u.Fragment.File.String()
		}
		// TODO: how to handle fragment params SepChar, right now assumes & like query
		if u.Fragment.Params != nil && len(u.Fragment.Params) > 0 {
			for i, p := range u.Fragment.Params {
				uri += p.String()
				if i+1 != len(u.Fragment.Params) {
					uri += "&"
				}
			}
		}
	*/
	return ""
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
	File   *Ident
	Params []*KeyValueExpr
}
