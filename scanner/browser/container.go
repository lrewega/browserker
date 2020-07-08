package browser

import (
	"context"
	"sync"
	"sync/atomic"

	"gitlab.com/browserker/browserk"
)

type readyWait struct {
	readyCh chan struct{}
	closed  bool
}

// Container for various browser events
type Container struct {
	messageLock   sync.RWMutex
	messages      map[string]*browserk.HTTPMessage
	loadRequestID string

	respReadyLock sync.RWMutex
	respReady     map[string]*readyWait

	requestCount int32

	storageLock   sync.RWMutex
	storageEvents []*browserk.StorageEvent

	consoleLock   sync.RWMutex
	consoleEvents []*browserk.ConsoleEvent
}

// NewContainer for holding request/responses, storage and console events
// Get*(s) methods will clear the underlying container type
func NewContainer() *Container {
	return &Container{
		messages:      make(map[string]*browserk.HTTPMessage),
		respReady:     make(map[string]*readyWait),
		storageEvents: make([]*browserk.StorageEvent, 0),
	}
}

// AddStorageEvent to the container
func (c *Container) AddStorageEvent(evt *browserk.StorageEvent) {
	c.storageLock.Lock()
	c.storageEvents = append(c.storageEvents, evt)
	c.storageLock.Unlock()
}

// GetStorageEvents and clear the container
func (c *Container) GetStorageEvents() []*browserk.StorageEvent {
	c.storageLock.Lock()
	evts := make([]*browserk.StorageEvent, len(c.storageEvents))
	copy(evts, c.storageEvents)
	c.storageEvents = make([]*browserk.StorageEvent, 0)
	c.storageLock.Unlock()
	return evts
}

// AddConsoleEvent to the container
func (c *Container) AddConsoleEvent(evt *browserk.ConsoleEvent) {
	c.consoleLock.Lock()
	c.consoleEvents = append(c.consoleEvents, evt)
	c.consoleLock.Unlock()
}

// GetConsoleEvents and clear the container
func (c *Container) GetConsoleEvents() []*browserk.ConsoleEvent {
	c.consoleLock.Lock()
	evts := make([]*browserk.ConsoleEvent, len(c.consoleEvents))
	copy(evts, c.consoleEvents)
	c.consoleEvents = make([]*browserk.ConsoleEvent, 0)
	c.consoleLock.Unlock()
	return evts
}

// SetLoadRequest uses the requestID of the *first* request as
// our key to return the httpresponse in GetResponses.
func (c *Container) SetLoadRequest(request *browserk.HTTPRequest) {
	c.messageLock.Lock()
	defer c.messageLock.Unlock()
	if c.loadRequestID != "" {
		return
	}
	c.loadRequestID = request.RequestId
}

// GetRequest via requestID
func (c *Container) GetRequest(requestID string) *browserk.HTTPRequest {
	c.messageLock.RLock()
	m, ok := c.messages[requestID]
	c.messageLock.RUnlock()
	if !ok {
		return nil
	}
	return m.Request
}

// GetResponse via requestID
func (c *Container) GetResponse(requestID string) *browserk.HTTPResponse {
	c.messageLock.RLock()
	m, ok := c.messages[requestID]
	c.messageLock.RUnlock()
	if !ok {
		return nil
	}
	return m.Response
}

// GetMessage via requestID
func (c *Container) GetMessage(requestID string) *browserk.HTTPMessage {
	c.messageLock.RLock()
	m, ok := c.messages[requestID]
	c.messageLock.RUnlock()
	if !ok {
		return nil
	}
	return m
}

// GetMessages returns all of the http req/resp messages and clears the container
func (c *Container) GetMessages() []*browserk.HTTPMessage {
	c.messageLock.Lock()
	defer c.messageLock.Unlock()
	msgs := make([]*browserk.HTTPMessage, len(c.messages))
	i := 0
	for _, v := range c.messages {
		if v == nil {
			continue
		}
		msgs[i] = v
		i++
	}
	c.messages = make(map[string]*browserk.HTTPMessage, 0)
	atomic.StoreInt32(&c.requestCount, 0)
	c.loadRequestID = ""
	return msgs
}

// AddRequest to our map of messages // TODO: handle redirects better
func (c *Container) AddRequest(request *browserk.HTTPRequest) {
	var exist bool
	msg := &browserk.HTTPMessage{}

	c.messageLock.Lock()
	if msg, exist = c.messages[request.RequestId]; !exist {
		msg = &browserk.HTTPMessage{}
	}

	msg.Request = request
	c.messages[request.RequestId] = msg
	c.messageLock.Unlock()
}

// AddResponse to our map of messages
func (c *Container) AddResponse(response *browserk.HTTPResponse) {
	var exist bool
	msg := &browserk.HTTPMessage{}
	c.messageLock.Lock()
	if msg, exist = c.messages[response.RequestId]; !exist {
		msg = &browserk.HTTPMessage{}
	}
	msg.Response = response
	c.messages[response.RequestId] = msg
	c.messageLock.Unlock()
}

// IncRequest count
func (c *Container) IncRequest() {
	atomic.AddInt32(&c.requestCount, 1)
}

// DecRequest count
func (c *Container) DecRequest() {
	atomic.AddInt32(&c.requestCount, -1)
}

// OpenRequestCount number of open requests observed 0 means we have everything
func (c *Container) OpenRequestCount() int32 {
	return atomic.LoadInt32(&c.requestCount)
}

// WaitFor if we have a readyCh for this request, if we don't, make the channel.
// If we do, it is already closed so we can return
func (c *Container) WaitFor(ctx context.Context, requestID string) error {
	var ready *readyWait
	var ok bool

	defer c.remove(requestID)

	c.respReadyLock.Lock()
	if ready, ok = c.respReady[requestID]; !ok {
		ready = &readyWait{readyCh: make(chan struct{}), closed: false}
		c.respReady[requestID] = ready
	}
	c.respReadyLock.Unlock()

	select {
	case <-ready.readyCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// BodyReady signals WaitFor that the response is done, and we can start reading the body
func (c *Container) BodyReady(requestID string) {
	c.respReadyLock.Lock()
	if _, ok := c.respReady[requestID]; !ok {
		c.respReady[requestID] = &readyWait{readyCh: make(chan struct{}), closed: false}
	}

	if !c.respReady[requestID].closed {
		c.respReady[requestID].closed = true
		close(c.respReady[requestID].readyCh)
	}
	c.respReadyLock.Unlock()
}

// remove the request from our ready map
func (c *Container) remove(requestID string) {
	c.respReadyLock.Lock()
	delete(c.respReady, requestID)
	c.respReadyLock.Unlock()
}
