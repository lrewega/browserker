package browserk

import "crypto/md5"

// revive:disable:var-naming

type PluginEventType int8

const (
	EvtDocumentRequest PluginEventType = iota
	EvtHTTPRequest
	EvtHTTPResponse
	EvtInterceptedHTTPRequest
	EvtInterceptedHTTPResponse
	EvtWebSocketRequest
	EvtWebSocketResponse
	EvtURL
	EvtJSResponse
	EvtStorage
	EvtCookie
	EvtConsole
)

type PluginEvent struct {
	ID        []byte
	Type      PluginEventType
	URL       string
	Nav       *Navigation
	BCtx      *Context
	EventData *PluginEventData
}

func (e *PluginEvent) Hash() []byte {
	if e.ID != nil {
		return e.ID
	}
	hash := md5.New()
	hash.Write([]byte{byte(e.Type)})
	hash.Write([]byte(e.URL))
	hash.Write(e.Nav.ID)

	e.ID = hash.Sum(nil)
	return e.ID
}

type PluginEventData struct {
	ID                      []byte
	HTTPRequest             *HTTPRequest
	HTTPResponse            *HTTPResponse
	InterceptedHTTPRequest  *InterceptedHTTPRequest
	InterceptedHTTPResponse *InterceptedHTTPResponse
	Storage                 *StorageEvent
	Cookie                  *Cookie
	Console                 *ConsoleEvent
}

func HTTPRequestPluginEvent(bctx *Context, URL string, nav *Navigation, request *HTTPRequest) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtHTTPRequest)
	evt.EventData = &PluginEventData{HTTPRequest: request}
	hash := md5.New()
	hash.Write([]byte(request.Type))
	hash.Write([]byte(request.Request.Method))
	hash.Write([]byte(request.Request.Url))
	hash.Write([]byte(request.Request.UrlFragment))

	if request.Request.HasPostData {
		hash.Write([]byte(request.Request.PostData))
	}

	evt.EventData.ID = hash.Sum(nil)
	return evt
}

func HTTPResponsePluginEvent(bctx *Context, URL string, nav *Navigation, response *HTTPResponse) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtHTTPResponse)
	evt.EventData = &PluginEventData{HTTPResponse: response}
	return evt
}

func InterceptedHTTPRequestPluginEvent(bctx *Context, URL string, nav *Navigation, request *InterceptedHTTPRequest) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtInterceptedHTTPRequest)
	evt.EventData = &PluginEventData{InterceptedHTTPRequest: request}
	return evt
}

func InterceptedHTTPResponsePluginEvent(bctx *Context, URL string, nav *Navigation, response *InterceptedHTTPResponse) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtInterceptedHTTPResponse)
	evt.EventData = &PluginEventData{InterceptedHTTPResponse: response}
	return evt
}

func StoragePluginEvent(bctx *Context, URL string, nav *Navigation, storage *StorageEvent) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtStorage)
	evt.EventData = &PluginEventData{Storage: storage}
	return evt
}

func CookiePluginEvent(bctx *Context, URL string, nav *Navigation, cookie *Cookie) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtCookie)
	evt.EventData = &PluginEventData{Cookie: cookie}
	return evt
}

func ConsolePluginEvent(bctx *Context, URL string, nav *Navigation, console *ConsoleEvent) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtCookie)
	evt.EventData = &PluginEventData{Console: console}
	return evt
}

func newPluginEvent(bctx *Context, URL string, nav *Navigation, eventType PluginEventType) *PluginEvent {
	return &PluginEvent{
		Type: eventType,
		URL:  URL,
		Nav:  nav,
		BCtx: bctx,
	}
}
