package iterator

import (
	"strings"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/parsers"
)

type Injection struct {
	Data *browserk.InjectionExpr
	Loc  browserk.InjectionLocation
}

type InjectionIterator struct {
	req          *browserk.HTTPRequest
	method       browserk.InjectionExpr
	uri          *injast.URI
	locs         []browserk.InjectionExpr
	currentInj   browserk.InjectionExpr
	currentIndex int
	invalidParse bool
}

// NewInjectionIter for iterating over requests in a navigation.
// TODO: should parse the requests injection points so we only have to do it once
func NewInjectionIter(req *browserk.HTTPRequest) *InjectionIterator {
	it := &InjectionIterator{
		req:  req,
		locs: make([]browserk.InjectionExpr, 0),
	}
	it.method = &injast.Ident{Name: req.Request.Method, NamePos: 0, Location: browserk.InjectMethod}
	it.locs = append(it.locs, it.method)
	it.parseURI()
	return it
}

// Request returns a copy of the http request for modification?
func (it *InjectionIterator) Request() *browserk.HTTPRequest {
	return it.req.Copy()
}

func (it *InjectionIterator) parseURI() {
	var err error
	if it.req == nil || it.req.Request == nil || it.req.Request.Url == "" {
		it.invalidParse = true
		return
	}

	// just in case
	if !strings.HasPrefix(it.req.Request.Url, "http") {
		it.invalidParse = true
		return
	}

	_, uri := SplitHost(it.req.Request.Url)
	uri += it.req.Request.UrlFragment
	p := &parsers.URIParser{}

	it.uri, err = p.Parse(uri)
	if err != nil {
		it.invalidParse = true
	}
	it.locs = append(it.locs, it.uri.Fields...)
}

func (it *InjectionIterator) Method() string {
	return it.method.String()
}

// URI Returns the entire parsed URI for injection
func (it *InjectionIterator) URI() *injast.URI {
	return it.uri
}

func (it *InjectionIterator) Path() string {
	return it.uri.PathOnly()
}

func (it *InjectionIterator) File() string {
	return it.uri.FileOnly()
}

func (it *InjectionIterator) Seek(index int) {
	if index >= len(it.locs) {
		it.currentInj = nil
		return
	}
	it.currentIndex = index
	it.currentInj = it.locs[index]
}

func (it *InjectionIterator) Next() {
	it.Seek(it.currentIndex + 1)
}

func (it *InjectionIterator) Expr() browserk.InjectionExpr {
	return it.currentInj
}

func (it *InjectionIterator) Key() (string, browserk.InjectionLocation) {
	v, ok := it.currentInj.(*injast.KeyValueExpr)
	if !ok {
		return "", 0
	}
	return v.Key.String(), v.Location
}

func (it *InjectionIterator) Value() (string, browserk.InjectionLocation) {
	v, ok := it.currentInj.(*injast.KeyValueExpr)
	if !ok {
		return it.currentInj.String(), it.currentInj.Loc()
	}
	return v.Value.String(), v.Location
}

func (it *InjectionIterator) Valid() bool {
	if it.invalidParse || it.currentInj == nil {
		return false
	}
	return true
}

func (it *InjectionIterator) Rewind() {
	it.currentIndex = 0
	it.Seek(it.currentIndex)
}

func SplitHost(u string) (string, string) {
	uriStart := 0
	slashCount := 0
	for i := 0; i < len(u); i++ {
		if u[i] == '/' {
			slashCount++
		}
		if slashCount == 3 {
			uriStart = i
			break
		}
	}
	return u[:uriStart], u[uriStart:]
}