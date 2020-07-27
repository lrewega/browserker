package storage_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
	"gitlab.com/browserker/scanner/plugin/plugintest"
	"gitlab.com/browserker/scanner/plugin/storage"
)

var leaser = browser.NewLocalLeaser()

func init() {
	//leaser.SetProxy("http://127.0.0.1:9000")
	leaser.SetHeadless()
}

type u struct {
	Username string `json:"username"`
}

func TestStorage(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()
	ctx := context.Background()

	called := false

	toTest := [...]plugintest.AttackTests{
		{
			FormHandler: nil,
			URL:         "http://localhost:%s/forms/localstorage.html",
		},
		{
			FormHandler: nil,
			URL:         "http://localhost:%s/forms/sessionstorage.html",
		},
	}

	pluginServicer := mock.MakeMockPluginServicer()
	plug := storage.New(pluginServicer)
	pluginServicer.Register(plug)

	for _, attackTest := range toTest {
		p, srv := plugintest.RunTestServer("../plugintest/testdata/forms", "/form/result", attackTest.FormHandler)
		defer srv.Shutdown(ctx)

		target := fmt.Sprintf(attackTest.URL, p)
		targetURL, _ := url.Parse(target)
		ctx := context.Background()
		bCtx := mock.MakeMockContext(ctx, targetURL)
		bCtx.FormHandler = crawler.NewCrawlerFormHandler(&browserk.DefaultFormValues)
		bCtx.Scope = scanner.NewScopeService(targetURL)
		bCtx.Reporter = mock.MakeMockReporter()
		bCtx.PluginServicer = pluginServicer

		pluginServicer.DispatchEventFn = func(evt *browserk.PluginEvent) {
			if evt.Type == browserk.EvtStorage {
				plug.OnEvent(evt)
				called = true
			}
		}
		_, port, _, err := plugintest.GetNewNavPaths(bCtx, pool, target)
		if err != nil {
			t.Fatalf("error getting new nav paths: %s\n", err)
		}

		if !called {
			t.Fatalf("attack was not successful: %s\n", target)
		}

		called = false
		pool.Return(ctx, port)
		srv.Shutdown(ctx)
	}
}
