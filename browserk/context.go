package browserk

import (
	"context"
	"math"
	"sync"

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

	jsBeforeLock    *sync.RWMutex
	jsBeforeHandler []JSHandler
	jsAfterLock     *sync.RWMutex
	jsAfterHandler  []JSHandler

	reqLock      *sync.RWMutex
	reqHandlers  []RequestHandler
	respLock     *sync.RWMutex
	respHandlers []ResponseHandler
	evtLock      *sync.RWMutex
	evtHandlers  []EventHandler
}

func NewContext(ctx context.Context, cancelFn context.CancelFunc) *Context {
	return &Context{
		Ctx:          ctx,
		CtxComplete:  cancelFn,
		jsBeforeLock: &sync.RWMutex{},
		jsAfterLock:  &sync.RWMutex{},
		reqLock:      &sync.RWMutex{},
		respLock:     &sync.RWMutex{},
		evtLock:      &sync.RWMutex{},
	}
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
		jsBeforeLock:    &sync.RWMutex{},
		jsBeforeHandler: c.jsBeforeHandler,
		jsAfterLock:     &sync.RWMutex{},
		jsAfterHandler:  c.jsAfterHandler,
		reqLock:         &sync.RWMutex{},
		reqHandlers:     c.reqHandlers,
		respLock:        &sync.RWMutex{},
		respHandlers:    c.respHandlers,
		evtLock:         &sync.RWMutex{},
		evtHandlers:     c.evtHandlers,
	}
}

// NextReq calls the next handler
func (c *Context) NextReq(browser Browser, i *InterceptedHTTPRequest) {
	c.reqLock.RLock()
	for reqIndex := int8(0); reqIndex < int8(len(c.reqHandlers)); reqIndex++ {
		c.reqHandlers[reqIndex](c, browser, i)
	}
	c.reqLock.RUnlock()
}

// AddReqHandler adds new request handlers
func (c *Context) AddReqHandler(i ...RequestHandler) {
	c.reqLock.Lock()
	if c.reqHandlers == nil {
		c.reqHandlers = make([]RequestHandler, 0)
	}
	c.reqHandlers = append(c.reqHandlers, i...)
	c.reqLock.Unlock()
}

// NextResp calls the next handler
func (c *Context) NextResp(browser Browser, i *InterceptedHTTPResponse) {
	c.respLock.RLock()
	for respIndex := int8(0); respIndex < int8(len(c.respHandlers)); respIndex++ {
		c.respHandlers[respIndex](c, browser, i)
	}
	c.respLock.RUnlock()
}

// AddRespHandler adds new request handlers
func (c *Context) AddRespHandler(i ...ResponseHandler) {
	c.respLock.Lock()
	if c.respHandlers == nil {
		c.respHandlers = make([]ResponseHandler, 0)
	}
	c.respHandlers = append(c.respHandlers, i...)
	c.respLock.Unlock()
}

// NextEvt calls the next handler
func (c *Context) NextEvt() {
	c.evtLock.RLock()
	for evtIndex := int8(0); evtIndex < int8(len(c.evtHandlers)); evtIndex++ {
		c.evtHandlers[evtIndex](c)
	}
	c.evtLock.RUnlock()
}

// AddEvtHandler adds new request handlers
func (c *Context) AddEvtHandler(i ...EventHandler) {
	c.evtLock.Lock()
	if c.evtHandlers == nil {
		c.evtHandlers = make([]EventHandler, 0)
	}
	c.evtHandlers = append(c.evtHandlers, i...)
	c.evtLock.Unlock()
}

// NextJSBefore calls the next handler
func (c *Context) NextJSBefore(browser Browser) {
	c.jsBeforeLock.RLock()
	for jsBeforeIndex := int8(0); jsBeforeIndex < int8(len(c.jsBeforeHandler)); jsBeforeIndex++ {
		c.jsBeforeHandler[jsBeforeIndex](c, browser)
	}
	c.jsBeforeLock.RUnlock()
}

// AddJSBeforeHandler adds new request handlers
func (c *Context) AddJSBeforeHandler(i ...JSHandler) {
	c.jsBeforeLock.Lock()
	if c.jsBeforeHandler == nil {
		c.jsBeforeHandler = make([]JSHandler, 0)
	}
	c.jsBeforeHandler = append(c.jsBeforeHandler, i...)
	c.jsBeforeLock.Unlock()
}

// NextJSAfter calls the next handler
func (c *Context) NextJSAfter(browser Browser) {
	c.jsAfterLock.RLock()
	for jsAfterIndex := int8(0); jsAfterIndex < int8(len(c.jsAfterHandler)); jsAfterIndex++ {
		c.jsAfterHandler[jsAfterIndex](c, browser)
	}
	c.jsAfterLock.RUnlock()
}

// AddJSAfterHandler adds new js handlers
func (c *Context) AddJSAfterHandler(i ...JSHandler) {
	c.jsAfterLock.Lock()
	if c.jsAfterHandler == nil {
		c.jsAfterHandler = make([]JSHandler, 0)
	}
	c.jsAfterHandler = append(c.jsAfterHandler, i...)
	c.jsAfterLock.Unlock()
}
