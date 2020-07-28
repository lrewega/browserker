package browserk

import "context"

// PluginServicer does what it says
type PluginServicer interface {
	Init(ctx context.Context) error
	Inject(mainContext *Context, injector Injector)
	Register(plugin Plugin)
	Unregister(plugin Plugin)
	DispatchEvent(evt *PluginEvent)
	RegisterForResponse(requestID string, respCh chan<- *InterceptedHTTPMessage, injection *InterceptedHTTPRequest)
	DispatchResponse(requestID string, interceptedMessage *InterceptedHTTPResponse)
	Store() PluginStorer
}
