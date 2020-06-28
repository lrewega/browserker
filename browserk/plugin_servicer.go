package browserk

import "context"

// PluginServicer does what it says
type PluginServicer interface {
	Name() string
	Init(ctx context.Context) error
	Inject(mainContext *Context, injector Injector)
	Register(plugin Plugin)
	Unregister(plugin Plugin)
	DispatchEvent(evt *PluginEvent)
	RegisterForResponse(requestID string, respCh chan<- *InterceptedHTTPResponse)
	DispatchResponse(requestID string, resp *InterceptedHTTPResponse)
	Store() PluginStorer
}
