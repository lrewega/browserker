package browser

import (
	"context"
	"sync"
	"sync/atomic"
)

type Response struct {
	RequestID string
	Body      string
}

type ResponseContainer struct {
	responsesLock sync.RWMutex
	loadRequest   string
	responses     map[string]*Response

	readyLock sync.RWMutex
	ready     map[string]chan struct{}

	requestCount int32
}

func NewResponseContainer() *ResponseContainer {
	return &ResponseContainer{
		responses: make(map[string]*Response),
		ready:     make(map[string]chan struct{}),
	}
}

// SetLoadRequest uses the requestID of the *first* request as
// our key to return the httpresponse in GetResponses.
func (c *ResponseContainer) SetLoadRequest(requestID string) {
	c.responsesLock.Lock()
	defer c.responsesLock.Unlock()
	if c.loadRequest != "" {
		return
	}
	c.loadRequest = requestID

}

// GetResponses returns the main load reseponse and all responses then clears the container
func (c *ResponseContainer) GetResponses() (*Response, []*Response) {
	var loadResponse *Response
	c.responsesLock.Lock()
	defer c.responsesLock.Unlock()
	r := make([]*Response, len(c.responses))
	i := 0
	for _, v := range c.responses {
		if v == nil {
			continue
		}
		r[i] = v
		if v.RequestID == c.loadRequest {
			loadResponse = v
		}
		i++
	}
	c.responses = make(map[string]*Response, 0)
	c.loadRequest = ""
	return loadResponse, r
}

// Add a response to our map
func (c *ResponseContainer) Add(response *Response) {
	c.responsesLock.Lock()
	c.responses[response.RequestID] = response
	c.responsesLock.Unlock()
}

func (c *ResponseContainer) IncRequest() {
	atomic.AddInt32(&c.requestCount, 1)
}

func (c *ResponseContainer) DecRequest() {
	atomic.AddInt32(&c.requestCount, -1)
}

func (c *ResponseContainer) GetRequests() int32 {
	return atomic.LoadInt32(&c.requestCount)
}

// WaitFor see if we have a readyCh for this request, if we don't, make the channel
// if we do, it is already closed so we can return
func (c *ResponseContainer) WaitFor(ctx context.Context, requestID string) error {
	var readyCh chan struct{}
	var ok bool

	defer c.remove(requestID)

	c.readyLock.Lock()
	if readyCh, ok = c.ready[requestID]; !ok {
		readyCh = make(chan struct{})
		c.ready[requestID] = readyCh
	}
	c.readyLock.Unlock()

	select {
	case <-readyCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// BodyReady signals WaitFor that the response is done, and we can start reading the body
func (c *ResponseContainer) BodyReady(requestID string) {
	c.readyLock.Lock()
	if _, ok := c.ready[requestID]; !ok {
		c.ready[requestID] = make(chan struct{})
	}
	close(c.ready[requestID])
	c.readyLock.Unlock()
}

// remove the request from our ready map
func (c *ResponseContainer) remove(requestID string) {
	c.readyLock.Lock()
	delete(c.ready, requestID)
	c.readyLock.Unlock()
}
