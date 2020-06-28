package browserk

import "time"

// HTTPMessage is the request/response pair
type HTTPMessage struct {
	RequestTime  time.Time
	Request      *HTTPRequest
	RequestMod   *HTTPModifiedRequest
	ResponseTime time.Time
	Response     *HTTPResponse
	ResponseMod  *HTTPModifiedResponse
}

func (m *HTTPMessage) Copy() *HTTPMessage {
	return &HTTPMessage{
		RequestTime:  m.RequestTime,
		Request:      m.Request.Copy(),
		RequestMod:   m.RequestMod.Copy(),
		ResponseTime: m.ResponseTime,
		Response:     m.Response.Copy(),
		ResponseMod:  m.ResponseMod.Copy(),
	}
}

// revive:disable:var-naming
func MessagesAfterRequestTime(m []*HTTPMessage, t time.Time) []*HTTPMessage {
	messages := make([]*HTTPMessage, 0)
	for _, v := range m {
		if t.After(v.RequestTime) {
			messages = append(messages, v)
		}
	}
	return messages
}

func MessagesAfterResponseTime(m []*HTTPMessage, t time.Time) []*HTTPMessage {
	messages := make([]*HTTPMessage, 0)
	for _, v := range m {
		if t.After(v.ResponseTime) {
			messages = append(messages, v)
		}
	}
	return messages
}

func MessagesBeforeRequestTime(m []*HTTPMessage, t time.Time) []*HTTPMessage {
	messages := make([]*HTTPMessage, 0)
	for _, v := range m {
		if t.Before(v.RequestTime) {
			messages = append(messages, v)
		}
	}
	return messages
}

func MessagesBeforeResponseTime(m []*HTTPMessage, t time.Time) []*HTTPMessage {
	messages := make([]*HTTPMessage, 0)
	for _, v := range m {
		if t.Before(v.ResponseTime) {
			messages = append(messages, v)
		}
	}
	return messages
}
