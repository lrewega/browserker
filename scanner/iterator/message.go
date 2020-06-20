package iterator

import (
	"gitlab.com/browserker/browserk"
)

type MessageIterator struct {
	nav          *browserk.NavigationWithResult
	currentMsg   *browserk.HTTPMessage
	currentIndex int
}

// NewMessageIter for iterating over requests in a navigation.
// TODO: should parse the requests injection points so we only have to do it once
func NewMessageIter(nav *browserk.NavigationWithResult) *MessageIterator {
	return &MessageIterator{
		nav: nav,
	}
}

func (it *MessageIterator) Seek(to int) {
	if it.nav == nil || it.nav.Result == nil ||
		it.nav.Result.Messages == nil || to >= len(it.nav.Result.Messages) {
		it.currentMsg = nil
		return
	}
	it.currentIndex = to
	it.currentMsg = it.nav.Result.Messages[it.currentIndex]
}

func (it *MessageIterator) Next() {
	it.currentIndex++
	it.Seek(it.currentIndex)
}

func (it *MessageIterator) Request() *browserk.HTTPRequest {
	return it.currentMsg.Request
}

func (it *MessageIterator) Response() *browserk.HTTPResponse {
	return it.currentMsg.Response
}

func (it *MessageIterator) Valid() bool {
	if it.currentMsg == nil {
		return false
	}
	return true
}

func (it *MessageIterator) Rewind() {
	it.currentIndex = 0
	it.Seek(0)
}
