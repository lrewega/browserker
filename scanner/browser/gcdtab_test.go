package browser_test

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/browser"
	"golang.org/x/net/context"
)

var leaser = browser.NewLocalLeaser()

func testServer() (string, *http.Server) {
	srv := &http.Server{Handler: http.FileServer(http.Dir("testdata/"))}
	testListener, _ := net.Listen("tcp", ":0")
	_, testServerPort, _ := net.SplitHostPort(testListener.Addr().String())
	//testServerAddr := fmt.Sprintf("http://localhost:%s/", testServerPort)
	go func() {
		if err := srv.Serve(testListener); err != http.ErrServerClosed {
			log.Fatalf("Serve(): %s", err)
		}
	}()

	return testServerPort, srv
}

func TestStartBrowsers(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()

	ctx := context.Background()
	target, _ := url.Parse("http://example.com")
	bCtx := mock.MakeMockContext(ctx, target)
	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	err = b.Navigate(ctx, "http://example.com")
	if err != nil {
		t.Fatalf("error getting url %s\n", err)
	}

	//msgs, _ := b.GetMessages()
	//spew.Dump(msgs)
}

func TestHookRequests(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()

	ctx := context.Background()
	target, _ := url.Parse("http://example.com")
	bCtx := mock.MakeMockContext(ctx, target)

	hook := func(c *browserk.Context, b browserk.Browser, i *browserk.InterceptedHTTPRequest) {
		t.Logf("inside hook!")
		i.Modified.Url = "http://example.com"
	}
	bCtx.AddReqHandler([]browserk.RequestHandler{hook}...)

	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	err = b.Navigate(ctx, "http://example.com")
	if err != nil {
		t.Fatalf("error getting url %s\n", err)
	}

	//msgs, _ := b.GetMessages()
	//spew.Dump(msgs)
}

func TestGetElements(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()

	ctx := context.Background()
	target, _ := url.Parse("https://angularjs.org")
	bCtx := mock.MakeMockContext(ctx, target)

	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	b.Navigate(ctx, "https://angularjs.org")

	ele, err := b.FindElements("form")
	if err != nil {
		t.Fatalf("error getting elements: %s\n", err)
	}
	if ele == nil {
		t.Fatalf("expected form elements")
	}
	//spew.Dump(ele)
}

func TestGcdWindows(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()
	ctx := context.Background()
	p, srv := testServer()
	defer srv.Shutdown(ctx)

	u := fmt.Sprintf("http://localhost:%s/window_main.html", p)
	target, _ := url.Parse(u)
	bCtx := mock.MakeMockContext(ctx, target)
	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	err = b.Navigate(ctx, u)
	if err != nil {
		t.Fatalf("error getting url %s\n", err)
	}
	msgs, _ := b.GetMessages()
	if msgs == nil {
		t.Fatalf("expected msgs")
	}
}

func TestBaseHref(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()
	ctx := context.Background()

	p, srv := testServer()
	defer srv.Shutdown(ctx)

	u := fmt.Sprintf("http://localhost:%s/basehref.html", p)
	target, _ := url.Parse(u)
	bCtx := mock.MakeMockContext(ctx, target)

	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	err = b.Navigate(ctx, u)
	if err != nil {
		t.Fatalf("error getting url %s\n", err)
	}
	eles, _ := b.FindElements("base")
	if eles == nil {
		t.Fatalf("expected eles")
	}
}

func TestFragment(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()
	ctx := context.Background()

	p, srv := testServer()
	defer srv.Shutdown(ctx)

	u := fmt.Sprintf("http://localhost:%s/basehref.html#/test", p)
	target, _ := url.Parse(u)
	bCtx := mock.MakeMockContext(ctx, target)

	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	err = b.Navigate(ctx, u)
	if err != nil {
		t.Fatalf("error getting url %s\n", err)
	}
	msg, _ := b.GetMessages()
	if msg[0].Request.Request.UrlFragment != "#/test" {
		t.Fatalf("expected fragment to include hash")
	}
}
