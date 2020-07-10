package store_test

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/store"
)

func TestReporting(t *testing.T) {
	result := mock.MakeMockResult([]byte{4, 5, 6})
	r := &browserk.Report{
		ID:          []byte{1, 2, 3},
		CheckID:     "1",
		CWE:         79,
		Description: "xss",
		Remediation: "don't have xss",
		Nav:         mock.MakeMockNavi([]byte{7, 8, 9}),
		Result:      result,
		NavResultID: nil,
		Evidence: &browserk.Evidence{
			ID:     []byte{123, 234},
			String: "some evidence",
		},
		Reported: time.Now(),
	}
	r.Hash()

	os.RemoveAll("testdata/uniq")
	p := store.NewPluginStore("testdata/uniq")
	if err := p.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer p.Close()

	p.AddReport(r)
	reports, err := p.GetReports()
	if err != nil {
		t.Fatalf("error getting report: %s\n", err)
	}
	// todo: better testing
	if len(reports) != 1 {
		t.Fatalf("expected 1 report")
	}

	if reports[0].CWE != 79 {
		t.Fatalf("wrong cwe")
	}

	if reports[0].Description != "don't have xss" {
		t.Fatalf("wrong description")
	}

}

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
