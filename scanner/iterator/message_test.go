package iterator_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/iterator"
)

func TestMessageIterator(t *testing.T) {
	expectedMsg := 3
	nav := mock.MakeMockNavi([]byte{1})
	res := mock.MakeMockResult([]byte{1})
	rIter := iterator.NewMessageIter(&browserk.NavigationWithResult{Result: res, Navigation: nav})
	i := 0
	for rIter.Rewind(); rIter.Valid(); rIter.Next() {
		t.Logf("%v\n", rIter.Request())
		i++
	}
	if i != expectedMsg {
		t.Fatalf("expected %d msgs got %d\n", expectedMsg, i)
	}

	// test empty
	rIter = iterator.NewMessageIter(&browserk.NavigationWithResult{})
	i = 0
	expectedMsg = 0
	for rIter.Rewind(); rIter.Valid(); rIter.Next() {
		t.Logf("%v\n", rIter.Request())
		i++
	}
	if i != expectedMsg {
		t.Fatalf("expected %d msgs got %d\n", expectedMsg, i)
	}
}
