package iterator_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/iterator"
)

func TestInjectionIter(t *testing.T) {

	msg := mock.MakeMockMessages()
	req := msg[0].Request
	req.Request.Url = "http://example:8080/some/path.js?x=1&y=2#/test" //

	it := iterator.NewInjectionIter(req)
	for it.Rewind(); it.Valid(); it.Next() {
		name, loc := it.Name()
		if loc == browserk.InjectQuery || loc == browserk.InjectFragment {
			key, _ := it.Key()
			val, _ := it.Value()
			t.Logf("key: %s , val: %s\n", key, val)
		}
		t.Logf("%s\n", name)
	}
	t.Fail()
}
