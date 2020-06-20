package mock

import "gitlab.com/browserker/browserk"

type Plugin struct {
	NameFn     func() string
	NameCalled bool

	IDFn     func() string
	IDCalled bool

	ConfigFn     func() *browserk.PluginConfig
	ConfigCalled bool

	InitContextFn     func(bctx *browserk.Context)
	InitContextCalled bool

	OptionsFn     func() *browserk.PluginOpts
	OptionsCalled bool

	ReadyFn     func(browser browserk.Browser) (bool, error) // ready for injection or whatever, ret true if injected
	ReadyCalled bool

	OnEventFn     func(evt *browserk.PluginEvent)
	OnEventCalled bool
}

func (p *Plugin) Name() string {
	p.NameCalled = true
	return p.NameFn()
}

func (p *Plugin) ID() string {
	p.IDCalled = true
	return p.IDFn()
}

func (p *Plugin) Config() *browserk.PluginConfig {
	p.ConfigCalled = true
	return p.ConfigFn()
}

func (p *Plugin) InitContext(bctx *browserk.Context) {
	p.InitContextCalled = true
	p.InitContextFn(bctx)
}

func (p *Plugin) Options() *browserk.PluginOpts {
	p.OptionsCalled = true
	return p.OptionsFn()
}

func (p *Plugin) Ready(browser browserk.Browser) (bool, error) {
	p.ReadyCalled = true
	return p.ReadyFn(browser)
}

func (p *Plugin) OnEvent(evt *browserk.PluginEvent) {
	p.OnEventCalled = true
	p.OnEventFn(evt)
}

func MakeMockPlugin() *Plugin {
	p := &Plugin{}

	p.NameFn = func() string {
		return "TestPlugin"
	}

	p.IDFn = func() string {
		return "BR-P-9999"
	}

	p.ConfigFn = func() *browserk.PluginConfig {
		return &browserk.PluginConfig{
			Class:  "",
			Plugin: "",
			ID:     9,
		}
	}

	p.InitContextFn = func(bctx *browserk.Context) {

	}

	p.OptionsFn = func() *browserk.PluginOpts {
		return &browserk.PluginOpts{
			IsolatedRequests: true,
			WriteResponses:   true,
			WriteRequests:    true,
			WriteJS:          true,
			ListenResponses:  true,
			ListenRequests:   true,
			ListenStorage:    true,
			ListenCookies:    true,
			ListenConsole:    true,
			ListenURL:        true,
			ListenJS:         true,
			ExecutionType:    browserk.ExecAlways,
			Injections:       nil,
		}
	}

	p.OnEventFn = func(evt *browserk.PluginEvent) {}
	return p
}
