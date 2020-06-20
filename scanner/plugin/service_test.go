package plugin_test

import (
	"context"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/plugin"
)

func TestInit(t *testing.T) {
	m := mock.MakeMockConfig()
	pluginStore := mock.MakeMockPluginStore()
	s := plugin.New(m, pluginStore)
	ctx := context.Background()
	if err := s.Init(ctx); err != nil {
		t.Fatalf("error initializing plugin service: %s\n", err)
	}
}

func TestDispatch(t *testing.T) {
	m := mock.MakeMockConfig()
	pluginStore := mock.MakeMockPluginStore()
	s := plugin.New(m, pluginStore)
	ctx := context.Background()
	if err := s.Init(ctx); err != nil {
		t.Fatalf("error initializing plugin service: %s\n", err)
	}
	cookies := mock.MakeMockCookies()

	mPlugin := mock.MakeMockPlugin()
	s.Register(mPlugin)

	for _, cookie := range cookies {
		s.DispatchEvent(browserk.CookiePluginEvent(nil, "test", nil, cookie))
	}

	if !mPlugin.OnEventCalled {
		t.Fatalf("error plugin OnEvent was never called")
	}

	// test unregister
	s.Unregister(mPlugin)
	mPlugin.OnEventCalled = false
	for _, cookie := range cookies {
		s.DispatchEvent(browserk.CookiePluginEvent(nil, "test", nil, cookie))
	}
	if mPlugin.OnEventCalled {
		t.Fatalf("plugin should not be called after unregistering")
	}

	mPlugin.OptionsFn = func() *browserk.PluginOpts {
		return &browserk.PluginOpts{
			IsolatedRequests: false,
			WriteResponses:   false,
			WriteRequests:    false,
			WriteJS:          false,
			ListenResponses:  false,
			ListenRequests:   false,
			ListenStorage:    false,
			ListenCookies:    false,
			ListenConsole:    false,
			ListenURL:        false,
			ListenJS:         false,
			ExecutionType:    0,
			Injections:       nil,
		}
	}
	s.Register(mPlugin)
	for _, cookie := range cookies {
		s.DispatchEvent(browserk.CookiePluginEvent(nil, "test", nil, cookie))
	}

	if mPlugin.OnEventCalled {
		t.Fatalf("plugin should not be called after if it's not set to listen")
	}
}
