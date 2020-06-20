package iterator

import (
	"strings"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/injections/injast"
	"gitlab.com/browserker/scanner/injections/parsers"
)

type Injection struct {
	Data *injast.Expr
	Loc  browserk.InjectionLocation
}

type InjectionIterator struct {
	req    *browserk.HTTPRequest
	method injast.Expr
	uri    *injast.URI
	// headers *injast.Headers
	// body *injast.Body
	currentInj      injast.Expr
	currentLoc      browserk.InjectionLocation
	currentLocIndex int
	invalidParse    bool
}

// NewMessageIter for iterating over requests in a navigation.
// TODO: should parse the requests injection points so we only have to do it once
func NewInjectionIter(req *browserk.HTTPRequest) *InjectionIterator {
	it := &InjectionIterator{
		req: req,
	}
	it.method = &injast.Ident{Name: req.Request.Method, NamePos: 0}
	it.parseURI()
	return it
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

	uri := stripHost(it.req.Request.Url)
	uri += it.req.Request.UrlFragment
	p := &parsers.URIParser{}

	it.uri, err = p.Parse(uri)
	if err != nil {
		it.invalidParse = true
	}
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

func (it *InjectionIterator) Next() {
	switch it.currentLoc {
	case browserk.InjectMethod:
		it.currentLoc = browserk.InjectPath
		it.currentLocIndex = 0
		it.currentInj = it.uri.Paths[0]
	case browserk.InjectPath:
		if it.currentLocIndex+1 >= len(it.uri.Paths) {
			it.currentLoc = browserk.InjectFile
			it.currentLocIndex = 0
			it.currentInj = it.uri.File
		} else {
			it.currentLocIndex++
			it.currentInj = it.uri.Paths[it.currentLocIndex]
		}
	case browserk.InjectFile:
		if it.uri.HasParams() {
			it.currentLoc = browserk.InjectQueryName
			it.currentLocIndex = 0
			it.currentInj = it.uri.Query.Params[0]
		} else if it.uri.HasFragmentPath() {
			it.skipToFragmentPath()
		} else if it.uri.HasFragmentParams() {
			it.skipToFragmentParams()
		} else {
			it.currentInj = nil
		}
	case browserk.InjectQueryName:
		if it.currentLocIndex+1 >= len(it.uri.Query.Params) {
			if it.uri.HasFragmentPath() {
				it.skipToFragmentPath()
			} else if it.uri.HasFragmentParams() {
				it.skipToFragmentParams()
			} else {
				it.currentInj = nil
			}
		} else {
			it.currentLocIndex++
			it.currentInj = it.uri.Query.Params[it.currentLocIndex]
		}
	case browserk.InjectFragmentPath:
		if it.currentLocIndex+1 >= len(it.uri.Fragment.Paths) {
			if it.uri.HasFragmentParams() {
				it.skipToFragmentParams()
			} else {
				it.currentInj = nil
			}
		} else {
			it.currentLocIndex++
			it.currentInj = it.uri.Fragment.Paths[it.currentLocIndex]
		}
	case browserk.InjectFragmentName:
		if it.currentLocIndex+1 >= len(it.uri.Fragment.Params) {
			it.currentInj = nil
		} else {
			it.currentLocIndex++
			it.currentInj = it.uri.Fragment.Params[it.currentLocIndex]
		}
	default:
		it.currentInj = nil
	}
}

func (it *InjectionIterator) skipToFragmentPath() {
	it.currentLoc = browserk.InjectFragmentPath
	it.currentLocIndex = 0
	it.currentInj = it.uri.Fragment.Paths[0]
}

func (it *InjectionIterator) skipToFragmentParams() {
	it.currentLoc = browserk.InjectFragmentName
	it.currentLocIndex = 0
	it.currentInj = it.uri.Fragment.Params[0]
}

func (it *InjectionIterator) Name() (string, browserk.InjectionLocation) {
	// TODO: handle index params
	switch it.currentLoc {
	case browserk.InjectQueryName, browserk.InjectFragmentName:
		k, _ := it.currentInj.(*injast.KeyValueExpr)
		return k.Key.String(), it.currentLoc
	case browserk.InjectQueryValue, browserk.InjectFragmentValue:
		v, _ := it.currentInj.(*injast.KeyValueExpr)
		return v.Value.String(), it.currentLoc
	}
	return it.currentInj.String(), it.currentLoc
}

func (it *InjectionIterator) Valid() bool {
	if it.invalidParse || it.currentInj == nil {
		return false
	}
	return true
}

func (it *InjectionIterator) Rewind() {
	it.currentLoc = browserk.InjectMethod
	it.currentInj = it.method
	it.currentLocIndex = 0
}

func stripHost(u string) string {
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
	return u[uriStart:]
}
