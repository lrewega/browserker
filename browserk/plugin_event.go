package browserk

import (
	"crypto/md5"
)

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
	ID         []byte
	Type       PluginEventType
	URL        string
	Nav        *Navigation
	BCtx       *Context
	EventData  *PluginEventData
	Uniqueness Unique
}

func (e *PluginEvent) Hash() []byte {
	if e.ID != nil {
		return e.ID
	}
	hash := md5.New()
	hash.Write([]byte{byte(e.Type)})
	hash.Write([]byte(e.URL))
	if e.Nav != nil {
		hash.Write(e.Nav.ID)
	}
	hash.Write(e.EventData.ID)
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

func (p *PluginEventData) Hash() []byte {
	if p.ID != nil {
		return p.ID
	}
	// TODO: implement better uniqueness
	hash := md5.New()
	if p.HTTPRequest != nil {
		hash.Write([]byte(p.HTTPRequest.Type))
		hash.Write([]byte(p.HTTPRequest.Request.Method))
		hash.Write([]byte(p.HTTPRequest.Request.Url))
		hash.Write([]byte(p.HTTPRequest.Request.UrlFragment))

		if p.HTTPRequest.Request.HasPostData {
			hash.Write([]byte(p.HTTPRequest.Request.PostData))
		}
	} else if p.HTTPResponse != nil {
		hash.Write([]byte(p.HTTPResponse.Type))
		hash.Write([]byte(p.HTTPResponse.Response.Protocol))
		hash.Write([]byte(p.HTTPResponse.Response.Url))

		if p.HTTPResponse.BodyHash != nil {
			hash.Write([]byte(p.HTTPResponse.BodyHash))
		}

	} else if p.InterceptedHTTPRequest != nil {
		hash.Write([]byte(p.InterceptedHTTPRequest.Modified.Method))
		hash.Write([]byte(p.InterceptedHTTPRequest.Modified.Url))
		hash.Write([]byte(p.InterceptedHTTPRequest.Modified.PostData))
	} else if p.InterceptedHTTPResponse != nil {
		hash.Write([]byte(p.InterceptedHTTPResponse.Modified.Body))
	} else if p.Storage != nil {
		var localStorage byte
		if p.Storage.IsLocalStorage {
			localStorage = 1
		}
		hash.Write([]byte{byte(p.Storage.Type)})
		hash.Write([]byte(p.Storage.Key))
		hash.Write([]byte{localStorage})
		hash.Write([]byte(p.Storage.SecurityOrigin))
	} else if p.Cookie != nil {
		hash.Write([]byte(p.Cookie.Name))
		hash.Write([]byte(p.Cookie.Domain))
		hash.Write([]byte(p.Cookie.Path))
		hash.Write([]byte(p.Cookie.SameSite))
		var secure byte
		if p.Cookie.Secure {
			secure = 1
		}
		hash.Write([]byte{secure})

		var httponly byte
		if p.Cookie.HTTPOnly {
			httponly = 1
		}
		hash.Write([]byte{httponly})
	} else if p.Console != nil {
		hash.Write([]byte(p.Console.Level))
		hash.Write([]byte{byte(p.Console.Line)})
		hash.Write([]byte{byte(p.Console.Column)})
		hash.Write([]byte(p.Console.Source))
	}
	p.ID = hash.Sum(nil)
	return p.ID
}
func HTTPRequestPluginEvent(bctx *Context, URL string, nav *Navigation, request *HTTPRequest) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtHTTPRequest)
	evt.EventData = &PluginEventData{HTTPRequest: request}
	evt.Hash()
	return evt
}

func HTTPResponsePluginEvent(bctx *Context, URL string, nav *Navigation, response *HTTPResponse) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtHTTPResponse)
	evt.EventData = &PluginEventData{HTTPResponse: response}
	evt.Hash()
	return evt
}

func InterceptedHTTPRequestPluginEvent(bctx *Context, URL string, nav *Navigation, request *InterceptedHTTPRequest) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtInterceptedHTTPRequest)
	evt.EventData = &PluginEventData{InterceptedHTTPRequest: request}
	evt.Hash()
	return evt
}

func InterceptedHTTPResponsePluginEvent(bctx *Context, URL string, nav *Navigation, response *InterceptedHTTPResponse) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtInterceptedHTTPResponse)
	evt.EventData = &PluginEventData{InterceptedHTTPResponse: response}
	evt.Hash()
	return evt
}

func StoragePluginEvent(bctx *Context, URL string, nav *Navigation, storage *StorageEvent) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtStorage)
	evt.EventData = &PluginEventData{Storage: storage}
	evt.Hash()
	return evt
}

func CookiePluginEvent(bctx *Context, URL string, nav *Navigation, cookie *Cookie) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtCookie)
	evt.EventData = &PluginEventData{Cookie: cookie}
	evt.Hash()
	return evt
}

func ConsolePluginEvent(bctx *Context, URL string, nav *Navigation, console *ConsoleEvent) *PluginEvent {
	evt := newPluginEvent(bctx, URL, nav, EvtConsole)
	evt.EventData = &PluginEventData{Console: console}
	evt.Hash()
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
