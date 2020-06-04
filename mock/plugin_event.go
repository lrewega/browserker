package mock

import "gitlab.com/browserker/browserk"

func MakeMockPluginEvent(url string, evt browserk.PluginEventType) *browserk.PluginEvent {
	switch evt {
	case browserk.EvtCookie:
		return MakeMockCookieEvent(url)
	case browserk.EvtConsole:
		return MakeMockConsoleEvent(url)
	case browserk.EvtHTTPRequest:
		return MakeMockHTTPRequestEvent(url)
	case browserk.EvtHTTPResponse:
		return MakeMockHTTPResponseEvent(url)
	case browserk.EvtInterceptedHTTPRequest:
		return MakeMockInterceptedHTTPRequestEvent(url)
	case browserk.EvtInterceptedHTTPResponse:
		return MakeMockInterceptedHTTPResponseEvent(url)
	case browserk.EvtJSResponse:
	case browserk.EvtStorage:
		return MakeMockStorageEvent(url)
	case browserk.EvtURL:
	case browserk.EvtWebSocketRequest:
	case browserk.EvtWebSocketResponse:
	}
	return nil
}

func MakeMockCookieEvent(url string) *browserk.PluginEvent {
	cookies := MakeMockCookies()
	nav := MakeMockNavi([]byte{1, 2, 3})
	return browserk.CookiePluginEvent(nil, url, nav, cookies[0])
}

func MakeMockConsoleEvent(url string) *browserk.PluginEvent {
	console := MakeMockConsole()
	nav := MakeMockNavi([]byte{1, 2, 3})
	return browserk.ConsolePluginEvent(nil, url, nav, console[0])
}

func MakeMockHTTPRequestEvent(url string) *browserk.PluginEvent {
	req := MakeMockMessages()
	req[0].Request.Hash()
	nav := MakeMockNavi([]byte{1, 2, 3})
	return browserk.HTTPRequestPluginEvent(nil, url, nav, req[0].Request)
}

func MakeMockHTTPResponseEvent(url string) *browserk.PluginEvent {
	req := MakeMockMessages()
	req[0].Response.Hash()
	nav := MakeMockNavi([]byte{1, 2, 3})
	return browserk.HTTPResponsePluginEvent(nil, url, nav, req[0].Response)
}

func MakeMockInterceptedHTTPRequestEvent(url string) *browserk.PluginEvent {
	return nil
}

func MakeMockInterceptedHTTPResponseEvent(url string) *browserk.PluginEvent {
	return nil
}

func MakeMockStorageEvent(url string) *browserk.PluginEvent {
	return nil
}
