package iterator

import (
	"strings"

	"github.com/rs/zerolog/log"
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
	body         *injast.Body
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
	it.parseBody()
	return it
}

// Request returns a copy of the http request for modification
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
		log.Warn().Err(err).Msg("failed to parse uri")
		it.invalidParse = true
	}
	log.Debug().Int("uri_inj_count", len(it.uri.Fields)).Msg("parsed URI for injection")
	for _, f := range it.uri.Fields {
		all := Collect(f)
		it.locs = append(it.locs, all...)
	}
	//it.locs = append(it.locs, it.uri.Fields...)
}

func (it *InjectionIterator) parseBody() {
	var err error
	if it.req == nil || it.req.Request == nil || it.req.Request.PostData == "" {
		return
	}

	p := &parsers.BodyParser{}
	it.body, err = p.Parse([]byte(it.req.Request.PostData))
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse postdata")
		return
	}
	log.Debug().Int("body_inj_count", len(it.body.Fields)).Msg("parsed body for injection")
	for _, f := range it.body.Fields {
		all := Collect(f)
		it.locs = append(it.locs, all...)
	}
	it.locs = append(it.locs, it.body.Fields...)
}

// Method of this request
func (it *InjectionIterator) Method() string {
	return it.method.String()
}

// URI Returns the entire parsed URI for injection
func (it *InjectionIterator) URI() *injast.URI {
	return it.uri
}

// SerializeURI or return the original if we failed to parse
func (it *InjectionIterator) SerializeURI() string {
	if it.uri != nil {
		return it.uri.String()
	}
	// if parsed wrong/bad or whatever just return the original
	_, uri := SplitHost(it.req.Request.Url)
	uri += it.req.Request.UrlFragment
	return uri
}

// Body returns the entire parsed Body for injection
func (it *InjectionIterator) Body() *injast.Body {
	return it.body
}

// SerializeBody or return the original if we failed to parse
func (it *InjectionIterator) SerializeBody() string {
	if it.body != nil {
		return it.body.String()
	}
	// if parsed wrong/bad or whatever just return the original
	return it.req.Request.PostData
}

// Path only for this URI
func (it *InjectionIterator) Path() string {
	return it.uri.PathOnly()
}

// File only for this URI
func (it *InjectionIterator) File() string {
	return it.uri.FileOnly()
}

// Seek to a specific injection index
func (it *InjectionIterator) Seek(index int) {
	if index >= len(it.locs) {
		it.currentInj = nil
		return
	}
	it.currentIndex = index
	it.currentInj = it.locs[index]
}

// Next injection expr, move iterator up one
func (it *InjectionIterator) Next() {
	it.Seek(it.currentIndex + 1)
}

// Expr is the current injection expr
func (it *InjectionIterator) Expr() browserk.InjectionExpr {
	return it.currentInj
}

// Key of the current injection expr
func (it *InjectionIterator) Key() (string, browserk.InjectionLocation) {
	v, ok := it.currentInj.(*injast.KeyValueExpr)
	if !ok {
		return "", 0
	}
	return v.Key.String(), v.Location
}

// Value of the current injection expr
func (it *InjectionIterator) Value() (string, browserk.InjectionLocation) {
	v, ok := it.currentInj.(*injast.KeyValueExpr)
	if !ok {
		return it.currentInj.String(), it.currentInj.Loc()
	}
	return v.Value.String(), v.Location
}

// Valid returns if we had issues parsing
func (it *InjectionIterator) Valid() bool {
	if it.invalidParse || it.currentInj == nil {
		return false
	}
	return true
}

// Rewind the iterator
func (it *InjectionIterator) Rewind() {
	it.currentIndex = 0
	it.Seek(it.currentIndex)
}

// SplitHost into host / uri
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
