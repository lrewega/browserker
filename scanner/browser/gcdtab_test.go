package browser_test

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/browser"
	"golang.org/x/net/context"
)

var leaser = browser.NewLocalLeaser()

func init() {
	leaser.SetHeadless()
}

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

	hook := func(c *browserk.Context, b browserk.Browser, i *browserk.InterceptedHTTPRequest) bool {
		t.Logf("inside hook!")
		i.Modified.Url = "http://example.com"
		return true
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

	ele, err := b.FindElements("form", true)
	if err != nil {
		t.Fatalf("error getting elements: %s\n", err)
	}
	if ele == nil {
		t.Fatalf("expected form elements")
	}

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
	eles, _ := b.FindElements("base", true)
	if eles == nil {
		t.Fatalf("expected eles")
	}
}

func TestInjectJS(t *testing.T) {
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
	largeStr := "console.log('" + strings.Repeat("A", 100000) + "');"
	action := &browserk.Action{
		Type:  browserk.ActExecuteJS,
		Input: []byte(largeStr),
	}

	if _, _, err := b.ExecuteAction(ctx, &browserk.Navigation{Action: action}); err != nil {
		t.Fatalf("error executing action: %s\n", err)
	}
	c := b.GetConsoleEvents()
	if len(c) != 1 {
		t.Fatalf("expected console events")
	}
	//spew.Dump(c)
	eles, _ := b.FindElements("base", true)
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

func TestInterceptLargeJS(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}

	defer leaser.Cleanup()
	ctx := context.Background()

	p, srv := testServer()
	defer srv.Shutdown(ctx)

	u := fmt.Sprintf("http://localhost:%s/bigjs.html", p)
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

	_, err = b.GetMessages()
	if err != nil {
		t.Fatalf("error getting messages %s\n", err)
	}
	//spew.Dump(m)
}

func TestInjectRequest(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}

	defer leaser.Cleanup()
	ctx := context.Background()

	p, srv := testServer()
	defer srv.Shutdown(ctx)

	u := fmt.Sprintf("http://localhost:%s/index.html", p)
	target, _ := url.Parse(u)
	bCtx := mock.MakeMockContext(ctx, target)
	respCh := make(chan *browserk.InterceptedHTTPMessage)

	bCtx.AddReqHandler(testInjectXHRReq(t, respCh, u+"?asdf=asdf", nil, "", []byte("")))

	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	err = b.Navigate(ctx, u)
	if err != nil {
		t.Fatalf("error getting url %s\n", err)
	}
	b.GetMessages()

	t.Logf("Injecting request...")
	if err := b.InjectRequest(ctx, "GET", "/someur.html"); err != nil {
		t.Fatalf("error injecting request: %s\n", err)
	}

	<-respCh
	t.Logf("GOT RESPONSE FROM CH:\n")
}

func testInjectXHRReq(t *testing.T, respCh chan *browserk.InterceptedHTTPMessage, newURI string, headers map[string]interface{}, body string, match []byte) browserk.RequestHandler {
	return func(bctx *browserk.Context, browser browserk.Browser, i *browserk.InterceptedHTTPRequest) bool {
		t.Logf("INTERCEPTED: %s = %s [%s]\n", i.RequestId, i.Request.Url, i.NetworkId)
		if !strings.HasSuffix(i.Request.Url, "someur.html") {
			return false
		}
		t.Logf("handling injection %s\n", i.Request.Url)
		i.Modified.Url = newURI
		i.Modified.SetHeaders(headers)
		i.Modified.PostData = body
		bctx.PluginServicer.RegisterForResponse(i.FrameId+i.NetworkId, respCh, i)
		return true
	}
}
