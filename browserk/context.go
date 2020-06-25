package browserk

import (
	"context"
	"math"

	"github.com/rs/zerolog"
)

// RequestHandler for adding middleware between browser HTTP Request events
type RequestHandler func(c *Context, browser Browser, i *InterceptedHTTPRequest)

// ResponseHandler for adding middleware between browser HTTP Response events
type ResponseHandler func(c *Context, browser Browser, i *InterceptedHTTPResponse)

// EventHandler for adding middleware between browser events
type EventHandler func(c *Context)

// JSHandler for adding middleware to exec JS before/after navigations
type JSHandler func(c *Context, browser Browser)

const abortIndex int8 = math.MaxInt8 / 2

// Context shared between services, browsers and plugins
type Context struct {
	Ctx            context.Context
	Log            *zerolog.Logger
	CtxComplete    func()
	Auth           AuthService
	Scope          ScopeService
	FormHandler    FormHandler
	Reporter       Reporter
	Crawl          CrawlGrapher
	PluginServicer PluginServicer

	jsBeforeHandler []JSHandler
	jsAfterHandler  []JSHandler
	reqHandlers     []RequestHandler
	respHandlers    []ResponseHandler
	evtHandlers     []EventHandler
}

// Copy the context services and handlers
func (c *Context) Copy() *Context {
	return &Context{
		Ctx:             c.Ctx,
		CtxComplete:     c.CtxComplete,
		Scope:           c.Scope,
		FormHandler:     c.FormHandler,
		Reporter:        c.Reporter,
		Crawl:           c.Crawl,
		PluginServicer:  c.PluginServicer,
		jsBeforeHandler: c.jsBeforeHandler,
		jsAfterHandler:  c.jsAfterHandler,
		reqHandlers:     c.reqHandlers,
		respHandlers:    c.respHandlers,
		evtHandlers:     c.evtHandlers,
	}
}

// NextReq calls the next handler
func (c *Context) NextReq(browser Browser, i *InterceptedHTTPRequest) {
	for reqIndex := int8(0); reqIndex < int8(len(c.reqHandlers)); reqIndex++ {
		c.reqHandlers[reqIndex](c, browser, i)
	}
}

// AddReqHandler adds new request handlers
func (c *Context) AddReqHandler(i ...RequestHandler) {
	if c.reqHandlers == nil {
		c.reqHandlers = make([]RequestHandler, 0)
	}
	c.reqHandlers = append(c.reqHandlers, i...)
}

// NextResp calls the next handler
func (c *Context) NextResp(browser Browser, i *InterceptedHTTPResponse) {
	for respIndex := int8(0); respIndex < int8(len(c.respHandlers)); respIndex++ {
		c.respHandlers[respIndex](c, browser, i)
	}
}

// AddRespHandler adds new request handlers
func (c *Context) AddRespHandler(i ...ResponseHandler) {
	if c.respHandlers == nil {
		c.respHandlers = make([]ResponseHandler, 0)
	}
	c.respHandlers = append(c.respHandlers, i...)
}

// NextEvt calls the next handler
func (c *Context) NextEvt() {
	for evtIndex := int8(0); evtIndex < int8(len(c.evtHandlers)); evtIndex++ {
		c.evtHandlers[evtIndex](c)
	}
}

// AddEvtHandler adds new request handlers
func (c *Context) AddEvtHandler(i ...EventHandler) {
	if c.evtHandlers == nil {
		c.evtHandlers = make([]EventHandler, 0)
	}
	c.evtHandlers = append(c.evtHandlers, i...)
}

// NextJSBefore calls the next handler
func (c *Context) NextJSBefore(browser Browser) {
	for jsBeforeIndex := int8(0); jsBeforeIndex < int8(len(c.jsBeforeHandler)); jsBeforeIndex++ {
		c.jsBeforeHandler[jsBeforeIndex](c, browser)
	}
}

// AddJSBeforeHandler adds new request handlers
func (c *Context) AddJSBeforeHandler(i ...JSHandler) {
	if c.jsBeforeHandler == nil {
		c.jsBeforeHandler = make([]JSHandler, 0)
	}
	c.jsBeforeHandler = append(c.jsBeforeHandler, i...)
}

// NextJSAfter calls the next handler
func (c *Context) NextJSAfter(browser Browser) {
	for jsAfterIndex := int8(0); jsAfterIndex < int8(len(c.jsAfterHandler)); jsAfterIndex++ {
		c.jsAfterHandler[jsAfterIndex](c, browser)
	}
}

// AddJSAfterHandler adds new js handlers
func (c *Context) AddJSAfterHandler(i ...JSHandler) {
	if c.jsAfterHandler == nil {
		c.jsAfterHandler = make([]JSHandler, 0)
	}
	c.jsAfterHandler = append(c.jsAfterHandler, i...)
}
