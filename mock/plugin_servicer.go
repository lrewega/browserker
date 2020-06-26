package mock

import (
	"context"
	"sync"

	"gitlab.com/browserker/browserk"
)

type PluginServicer struct {
	InitFn     func(ctx context.Context) error
	InitCalled bool

	RegisterFn     func(plugin browserk.Plugin)
	RegisterCalled bool

	UnregisterFn     func(plugin browserk.Plugin)
	UnregisterCalled bool

	DispatchEventFn     func(evt *browserk.PluginEvent)
	DispatchEventCalled bool

	StoreFn     func() browserk.PluginStorer
	StoreCalled bool

	RegisterForResponseFn     func(requestID string, respCh chan<- *browserk.InterceptedHTTPResponse)
	RegisterForResponseCalled bool

	DispatchResponseFn     func(requestID string, resp *browserk.InterceptedHTTPResponse)
	DispatchResponseCalled bool
}

func (p *PluginServicer) Name() string {
	return "PluginService"
}

func (p *PluginServicer) Init(ctx context.Context) error {
	p.InitCalled = true
	return p.InitFn(ctx)
}
func (p *PluginServicer) Register(plugin browserk.Plugin) {
	p.RegisterCalled = true
	p.RegisterFn(plugin)
}

func (p *PluginServicer) Unregister(plugin browserk.Plugin) {
	p.UnregisterCalled = true
	p.UnregisterFn(plugin)
}

func (p *PluginServicer) DispatchEvent(evt *browserk.PluginEvent) {
	p.DispatchEventCalled = true
	p.DispatchEventFn(evt)
}

func (p *PluginServicer) Store() browserk.PluginStorer {
	p.StoreCalled = true
	return p.StoreFn()
}

func (p *PluginServicer) RegisterForResponse(requestID string, respCh chan<- *browserk.InterceptedHTTPResponse) {
	p.RegisterForResponseCalled = true
	p.RegisterForResponseFn(requestID, respCh)
}

func (p *PluginServicer) DispatchResponse(requestID string, resp *browserk.InterceptedHTTPResponse) {
	p.DispatchResponseCalled = true
	p.DispatchResponseFn(requestID, resp)
}

func MakeMockPluginServicer() *PluginServicer {
	p := &PluginServicer{}
	p.InitFn = func(ctx context.Context) error {
		return nil
	}

	plugins := make(map[string]browserk.Plugin)
	resps := make(map[string]chan<- *browserk.InterceptedHTTPResponse)
	pLock := &sync.RWMutex{}

	p.RegisterFn = func(plugin browserk.Plugin) {
		pLock.Lock()
		defer pLock.Unlock()
		plugins[plugin.ID()] = plugin
	}

	p.UnregisterFn = func(plugin browserk.Plugin) {
		pLock.Lock()
		defer pLock.Unlock()
		delete(plugins, plugin.ID())
	}

	p.DispatchEventFn = func(evt *browserk.PluginEvent) {
		pLock.RLock()
		defer pLock.RUnlock()
		for _, p := range plugins {
			p.OnEvent(evt)
		}
	}

	p.DispatchResponseFn = func(requestID string, resp *browserk.InterceptedHTTPResponse) {
		pLock.RLock()
		defer pLock.RUnlock()
		if respCh, ok := resps[requestID]; ok {
			delete(resps, requestID)
			respCh <- resp.Copy()
		}
	}

	p.RegisterForResponseFn = func(requestID string, respCh chan<- *browserk.InterceptedHTTPResponse) {
		pLock.RLock()
		defer pLock.RUnlock()
		resps[requestID] = respCh
	}

	p.StoreFn = func() browserk.PluginStorer {
		return nil
	}

	return p
}
