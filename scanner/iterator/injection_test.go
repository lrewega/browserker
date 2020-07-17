package iterator_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/iterator"
)

func TestInjectionIter(t *testing.T) {

	msg := mock.MakeMockMessages()
	req := msg[0].Request
	req.Request.Url = "http://localhost/vulnerabilities/fi/?page=include.php" //

	it := iterator.NewInjectionIter(req)
	for it.Rewind(); it.Valid(); it.Next() {
		name, loc := it.Value()
		if loc == browserk.InjectQuery || loc == browserk.InjectFragment {
			key, _ := it.Key()
			val, _ := it.Value()
			if val == "1" {
				it.Expr().Inject("xss", browserk.InjectValue)
			}
			t.Logf("key: %s , val: %s\n", key, val)
			if it.URI().String() != "/vulnerabilities/fi/?page=xss" {
				t.Fatalf("failed to inject value")
			}
		}
		t.Logf("%s\n", name)
	}
}

func TestInjectionWithBodyIter(t *testing.T) {

	msg := mock.MakeMockMessages()
	req := msg[0].Request
	req.Request.Url = "http://example:8080/some" //
	req.Request.PostData = "x=1&y=2"

	it := iterator.NewInjectionIter(req)
	for it.Rewind(); it.Valid(); it.Next() {
		name, loc := it.Value()
		spew.Dump(loc)
		if loc == browserk.InjectBody {
			key, _ := it.Key()
			val, _ := it.Value()
			if val == "1" {
				it.Expr().Inject("xss", browserk.InjectValue)
			}
			t.Logf("key: %s , val: %s\n", key, val)
			if it.Body().String() != "x=xss&y=2" {
				t.Fatalf("failed to inject value")
			}
		}
		t.Logf("%s\n", name)
	}
}

func TestInjectionWithJSONBodyIter(t *testing.T) {

	msg := mock.MakeMockMessages()
	req := msg[0].Request
	req.Request.Url = "http://example:8080/some" //
	req.Request.PostData = "{\"x\": \"hi\"}"

	it := iterator.NewInjectionIter(req)
	for it.Rewind(); it.Valid(); it.Next() {
		name, loc := it.Value()
		spew.Dump(loc)
		if loc == browserk.InjectBody {
			key, _ := it.Key()
			val, _ := it.Value()
			if val == "hi" {
				it.Expr().Inject("xs\"s", browserk.InjectValue)
			}
			t.Logf("key: %s , val: %s\n", key, val)
			if it.Body().String() != "{\"x\": \"xs\\\"s\"}" {
				t.Fatalf("failed to inject value")
			}
		}
		t.Logf("%s\n", name)
	}
}

func TestSplitHost(t *testing.T) {
	host, uri := iterator.SplitHost("http://example:8080/some/path.js?x=1&y=2#/test")
	if host != "http://example:8080" {
		t.Fatalf("expected %s got %s", "http://example:8080", host)
	}
	if uri != "/some/path.js?x=1&y=2#/test" {
		t.Fatalf("expected %s got %s", "/some/path.js?x=1&y=2#/test", uri)
	}
}
