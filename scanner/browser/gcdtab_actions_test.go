package browser_test

import (
	"fmt"
	"net/url"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/browser"
	"golang.org/x/net/context"
)

func TestActionClick(t *testing.T) {
	leaser.SetHeadless()
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}

	defer leaser.Cleanup()

	ctx := context.Background()
	p, srv := testServer()
	defer srv.Shutdown(ctx)

	u := fmt.Sprintf("http://localhost:%s/events.html", p)
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
	eles, err := b.FindElements(ctx, "button", true)
	if err != nil {
		t.Fatalf("error getting elements: %s\n", err)
	}

	for _, ele := range eles {
		nav := &browserk.Navigation{}
		act := &browserk.Action{
			Type:    browserk.ActLeftClick,
			Element: ele,
		}
		nav.Action = act

		_, causedLoad, err := b.ExecuteAction(ctx, nav)
		if err != nil {
			t.Fatalf("error executing click: %s\n", err)
		}

		if causedLoad {
			t.Fatalf("load should not have been caused")
		}
	}

	evts := b.GetConsoleEvents()
	if evts == nil {
		t.Fatalf("console events were not captured")
	}
	if len(evts) != 3 {
		t.Fatalf("expected 3 console log events, got %d\n", len(evts))
	}
}
