package iterator_test

import (
	"testing"

	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/iterator"
)

func TestInjectionIter(t *testing.T) {

	msg := mock.MakeMockMessages()
	req := msg[0].Request
	req.Request.Url = "http://example:8080/some/path?x=1&y=2#/test"

	it := iterator.NewInjectionIter(req)
	for it.Rewind(); it.Valid(); it.Next() {
		name, loc := it.Name()
		t.Logf("%s %d\n", name, loc)
	}
	t.Fail()
}
