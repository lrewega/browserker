package browserk

import (
	"context"
	"math"
)

// RequestHandler for adding middleware between browser HTTP Request events
type RequestHandler func(c *Context)

// ResponseHandler for adding middleware between browser HTTP Response events
type ResponseHandler func(c *Context)

// EventHandler for adding middleware between browser events
type EventHandler func(c *Context)

// JSHandler for adding middleware between browser events
type JSHandler func(c *Context)

const abortIndex int8 = math.MaxInt8 / 2

// Context shared between services, browsers and plugins
type Context struct {
	Ctx      *context.Context
	Scope    ScopeService
	Reporter Reporter
	Injector Injector
	Crawl    CrawlGrapher
	Attack   AttackGrapher

	jsBeforeHandler []JSHandler
	jsBeforeIndex   int8

	jsAfterHandler []JSHandler
	jsAfterIndex   int8

	reqHandlers []RequestHandler
	reqIndex    int8

	respHandlers []ResponseHandler
	respIndex    int8

	evtHandlers []EventHandler
	evtIndex    int8
}

func (c *Context) Copy() *Context {
	return &Context{
		Ctx:             c.Ctx,
		Scope:           c.Scope,
		Reporter:        c.Reporter,
		Injector:        c.Injector,
		Crawl:           c.Crawl,
		Attack:          c.Attack,
		jsBeforeHandler: c.jsBeforeHandler,
		jsBeforeIndex:   0,
		jsAfterHandler:  c.jsAfterHandler,
		jsAfterIndex:    0,
		reqHandlers:     c.reqHandlers,
		reqIndex:        0,
		respHandlers:    c.respHandlers,
		respIndex:       0,
		evtHandlers:     c.evtHandlers,
		evtIndex:        0,
	}
}

// NextReq calls the next handler
func (c *Context) NextReq() {
	for c.reqIndex < int8(len(c.reqHandlers)) {
		c.reqHandlers[c.reqIndex](c)
		c.reqIndex++
	}
}

// AddReqHandler adds new request handlers
func (c *Context) AddReqHandler(i ...RequestHandler) {
	if c.reqHandlers == nil {
		c.reqHandlers = make([]RequestHandler, 0)
	}
	c.reqHandlers = append(c.reqHandlers, i...)
}

// IsReqAborted returns true if the current context was aborted.
func (c *Context) IsReqAborted() bool {
	return c.reqIndex >= abortIndex
}

// ReqAbort prevents pending handlers from being called. Call ReqAbort to ensure the remaining handlers
// for this request are not called.
func (c *Context) ReqAbort() {
	c.reqIndex = abortIndex
}

// NextResp calls the next handler
func (c *Context) NextResp() {
	for c.respIndex < int8(len(c.respHandlers)) {
		c.respHandlers[c.respIndex](c)
		c.respIndex++
	}

}

// AddRespHandler adds new request handlers
func (c *Context) AddRespHandler(i ...ResponseHandler) {
	if c.respHandlers == nil {
		c.respHandlers = make([]ResponseHandler, 0)
	}
	c.respHandlers = append(c.respHandlers, i...)
}

// IsRespAborted returns true if the current context was aborted.
func (c *Context) IsRespAborted() bool {
	return c.respIndex >= abortIndex
}

// RespAbort prevents pending handlers from being called. Call ReqAbort to ensure the remaining handlers
// for this request are not called.
func (c *Context) RespAbort() {
	c.respIndex = abortIndex
}

// NextEvt calls the next handler
func (c *Context) NextEvt() {
	for c.evtIndex < int8(len(c.evtHandlers)) {
		c.evtHandlers[c.evtIndex](c)
		c.evtIndex++
	}
}

// AddEvtHandler adds new request handlers
func (c *Context) AddEvtHandler(i ...EventHandler) {
	if c.evtHandlers == nil {
		c.evtHandlers = make([]EventHandler, 0)
	}
	c.evtHandlers = append(c.evtHandlers, i...)
}

// IsEvtAborted returns true if the current context was aborted.
func (c *Context) IsEvtAborted() bool {
	return c.evtIndex >= abortIndex
}

// EvtAbort prevents pending handlers from being called. Call ReqAbort to ensure the remaining handlers
// for this request are not called.
func (c *Context) EvtAbort() {
	c.evtIndex = abortIndex
}

// NextJSBefore calls the next handler
func (c *Context) NextJSBefore() {
	for c.jsBeforeIndex < int8(len(c.jsBeforeHandler)) {
		c.jsBeforeHandler[c.jsBeforeIndex](c)
		c.jsBeforeIndex++
	}
}

// AddJSBeforeHandler adds new request handlers
func (c *Context) AddJSBeforeHandler(i ...JSHandler) {
	if c.jsBeforeHandler == nil {
		c.jsBeforeHandler = make([]JSHandler, 0)
	}
	c.jsBeforeHandler = append(c.jsBeforeHandler, i...)
}

// IsJSBeforeAborted returns true if the current context was aborted.
func (c *Context) IsJSBeforeAborted() bool {
	return c.jsBeforeIndex >= abortIndex
}

// JSBeforeAbort prevents pending handlers from being called. Call ReqAbort to ensure the remaining handlers
// for this request are not called.
func (c *Context) JSBeforeAbort() {
	c.jsBeforeIndex = abortIndex
}

// NextJSAfter calls the next handler
func (c *Context) NextJSAfter() {
	for c.jsAfterIndex < int8(len(c.jsAfterHandler)) {
		c.jsAfterHandler[c.jsAfterIndex](c)
		c.jsBeforeIndex++
	}
}

// AddJSAfterHandler adds new request handlers
func (c *Context) AddJSAfterHandler(i ...JSHandler) {
	if c.jsAfterHandler == nil {
		c.jsAfterHandler = make([]JSHandler, 0)
	}
	c.jsAfterHandler = append(c.jsAfterHandler, i...)
}

// IsJSAfterAborted returns true if the current context was aborted.
func (c *Context) IsJSAfterAborted() bool {
	return c.jsAfterIndex >= abortIndex
}

// JSAfterAbort prevents pending handlers from being called. Call ReqAbort to ensure the remaining handlers
// for this request are not called.
func (c *Context) JSAfterAbort() {
	c.jsAfterIndex = abortIndex
}
