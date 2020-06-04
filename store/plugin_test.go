package store_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/store"
)

func TestUnique(t *testing.T) {
	os.RemoveAll("testdata/uniq")
	p := store.NewPluginStore("testdata/uniq")
	if err := p.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer p.Close()

	target, _ := url.Parse("https://example.com/")
	bctx := mock.MakeMockContext(context.Background(), target)

	evt := mock.MakeMockPluginEvent("https://example.com/some/bloody/path?x=1", browserk.EvtCookie)
	evt.BCtx = bctx

	u := p.IsUnique(evt)
	testAllUnique(u, t)
	u = p.IsUnique(evt)
	testAllNotUnique(u, t)

	evt = mock.MakeMockPluginEvent("https://example.com/some/bloody/path?x=1&y=2", browserk.EvtCookie)
	evt.BCtx = bctx
	u = p.IsUnique(evt)
	if u.File() {
		t.Fatalf("expected File to not be unique\n")
	}
	if !u.Query() {
		t.Fatalf("expected Query to be unique\n")
	}
	if !u.Fragment() {
		t.Fatalf("expected Fragment to be unique\n")
	}
}
func testAllUnique(u browserk.Unique, t *testing.T) {
	if !u.Host() {
		t.Fatalf("expected Host to be unique\n")
	}
	if !u.Path() {
		t.Fatalf("expected Path to be unique\n")
	}
	if !u.File() {
		t.Fatalf("expected File to be unique\n")
	}
	if !u.Query() {
		t.Fatalf("expected Query to be unique\n")
	}
	if !u.Fragment() {
		t.Fatalf("expected Fragment to be unique\n")
	}
}

func testAllNotUnique(u browserk.Unique, t *testing.T) {
	if u.Host() {
		t.Fatalf("expected Host to not be unique\n")
	}
	if u.Path() {
		t.Fatalf("expected Path to not be unique\n")
	}
	if u.File() {
		t.Fatalf("expected File to not be unique\n")
	}
	if u.Query() {
		t.Fatalf("expected Query to not be unique\n")
	}
	if u.Fragment() {
		t.Fatalf("expected Fragment to not be unique\n")
	}
}
