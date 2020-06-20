package plugin

import (
	"sync"

	"gitlab.com/browserker/browserk"
)

// Container for concurrent safe access and execution
type Container struct {
	lock    *sync.RWMutex
	plugins map[string]browserk.Plugin
}

// NewContainer for plugins
func NewContainer() *Container {
	return &Container{
		lock:    &sync.RWMutex{},
		plugins: make(map[string]browserk.Plugin),
	}
}

// Add a plugin from our container
func (c *Container) Add(plugin browserk.Plugin) {
	c.lock.Lock()
	c.plugins[plugin.ID()] = plugin
	c.lock.Unlock()
}

// Remove a plugin from our container
func (c *Container) Remove(plugin browserk.Plugin) {
	c.lock.Lock()
	delete(c.plugins, plugin.ID())
	c.lock.Unlock()
}

// Call a plugin if the event type matches the options provided by a plugin
func (c *Container) Call(evt *browserk.PluginEvent) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, plugin := range c.plugins {
		if evt.Type == browserk.EvtHTTPRequest && plugin.Options().ListenRequests {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtHTTPResponse && plugin.Options().ListenResponses {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtInterceptedHTTPRequest && plugin.Options().ListenRequests {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtInterceptedHTTPResponse && plugin.Options().ListenResponses {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtWebSocketRequest && plugin.Options().ListenRequests {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtWebSocketResponse && plugin.Options().ListenResponses {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtURL && plugin.Options().ListenURL {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtJSResponse && plugin.Options().ListenJS {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtStorage && plugin.Options().ListenStorage {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtCookie && plugin.Options().ListenCookies {
			plugin.OnEvent(evt)
		} else if evt.Type == browserk.EvtConsole && plugin.Options().ListenConsole {
			plugin.OnEvent(evt)
		}
	}
}
