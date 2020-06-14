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
			u.updatePos(&lastPos, p)
			// if a path has been sliced out, skip it but update our lastPos
			if p.Modded && p.Mod == "" {
				continue
			}
			uri += p.String() + "/"

			lastPos++ // add 1 for slash
		}
	}

	u.updatePos(&lastPos, u.File)
	if u.File != nil && u.File.String() != "" {
		// only add the file if it wasn't modified, or the Mod is not an empty string
		if !u.File.Modded || u.File.Mod != "" {
			uri += u.File.String()
		}
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
			u.updatePos(&lastPos, p)
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
			u.updatePos(&lastPos, p)
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
			u.updatePos(&lastPos, p)
		}
	}

	// account for trailing delimiters
	if int(lastPos) != len(u.Original) {
		uri += string(u.Original[lastPos:len(u.Original)])
	}

	u.Modified = []byte(uri)
	return uri
}

// make sure Pos isn't 0 as injected nodes will have a 0 pos.
func (u *URI) updatePos(pos *Pos, n Node) {
	if n.Pos() == 0 {
		return
	}
	*pos = n.End()
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

// ReplaceFile with newFile name
func (u *URI) ReplaceFile(newFile string) {
	if u.File == nil {
		u.File = &Ident{Name: ""}
	}
	u.File.Modify(newFile)
}

// ReplacePath with a new string
func (u *URI) ReplacePath(path string, index int) bool {
	return u.replaceIdent(u.Paths, path, index)
}

func (u *URI) ReplaceFragmentPath(path string, index int) bool {
	return u.replaceIdent(u.Fragment.Paths, path, index)
}

func (u *URI) replaceIdent(idents []*Ident, path string, index int) bool {
	if index < 0 || index > len(idents) {
		return false
	}
	idents[index].Modify(path)
	return true
}

// ReplaceParam finds the first occurance of original (as the key) and replaces it
// with newKey and newVal
func (u *URI) ReplaceParam(original, newKey, newVal string) bool {
	return u.replaceParam(u.Query.Params, original, newKey, newVal)
}

// ReplaceFragmentParam finds the first occurance of original (as the key) and replaces it
// with newKey and newVal for fragments
func (u *URI) ReplaceFragmentParam(original, newKey, newVal string) bool {
	return u.replaceParam(u.Fragment.Params, original, newKey, newVal)
}

func (u *URI) replaceParam(params []*KeyValueExpr, original, newKey, newVal string) bool {
	for _, kv := range params {
		if kv.Key.String() == original {
			keyMod := true
			if original != newKey {
				keyMod = ReplaceExpr(kv, newKey, "key")
			}
			valMod := ReplaceExpr(kv, newVal, "value")
			return keyMod && valMod
		}
	}
	return false
}

// ReplaceParamByIndex attempts to directly access the query param by index and replace
// instead of looking up the name
func (u *URI) ReplaceParamByIndex(index int, newKey, newVal string) bool {
	return u.replaceParamByIndex(u.Query.Params, index, newKey, newVal)
}

// ReplaceFragmentParamByIndex attempts to directly access the query param by index and replace
// instead of looking up the name for fragments
func (u *URI) ReplaceFragmentParamByIndex(index int, newKey, newVal string) bool {
	return u.replaceParamByIndex(u.Fragment.Params, index, newKey, newVal)
}

func (u *URI) replaceParamByIndex(params []*KeyValueExpr, index int, newKey, newVal string) bool {
	if index < 0 || index > len(params) {
		return false
	}

	kv := params[index]
	keyMod := ReplaceExpr(kv, newKey, "key")
	valMod := ReplaceExpr(kv, newVal, "value")
	return keyMod && valMod
}

// ReplaceIndexedParam replaces x[original]=1 with x[new]=1
func (u *URI) ReplaceIndexedParam(original, newKey, newIndexVal, newVal string) bool {
	return u.replaceIndexedParam(u.Query.Params, original, newKey, newIndexVal, newVal)
}

// ReplaceFragmentIndexedParam replaces x[original]=1 with x[new]=1
func (u *URI) ReplaceFragmentIndexedParam(original, newKey, newIndexVal, newVal string) bool {
	return u.replaceIndexedParam(u.Fragment.Params, original, newKey, newIndexVal, newVal)
}

func (u *URI) replaceIndexedParam(params []*KeyValueExpr, original, newKey, newIndexVal, newVal string) bool {
	for _, kv := range params {
		if kv.Key.String() == original {
			keyMod := true
			if original != newKey {
				keyMod = ReplaceExpr(kv, newKey, "key")
			}
			indexMod := ReplaceExpr(kv.Key, newIndexVal, "index")
			valMod := ReplaceExpr(kv, newVal, "value")
			return keyMod && valMod && indexMod
		}
	}
	return false
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
