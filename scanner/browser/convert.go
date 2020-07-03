package browser

import (
	"crypto/sha1"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wirepair/gcd/gcdapi"
	"gitlab.com/browserker/browserk"
)

// ElementToHTMLElement convert
func ElementToHTMLElement(ele *Element) *browserk.HTMLElement {
	var ok bool
	b := &browserk.HTMLElement{Events: make(map[string]browserk.HTMLEventType, 0)}
	b.Type = browserk.CUSTOM

	ele.WaitForReady()
	tag, err := ele.GetTagName()
	if err != nil {
		ele.tab.ctx.Log.Error().Err(err).Msg("getting tag name failed")
	}
	b.Type, ok = browserk.HTMLTypeMap[strings.ToUpper(tag)]
	if !ok {
		b.CustomTagName = tag
	}
	// if we can get the elements loc, it's not hidden :>
	switch b.Type {
	case browserk.HTML, browserk.SCRIPT, browserk.TITLE, browserk.NOSCRIPT, browserk.HEAD:
		b.Hidden = true
	default:
		_, _, err = ele.getCenter()
		if err != nil {
			//ele.tab.ctx.Log.Error().Err(err).Str("type", browserk.HTMLTypeToStrMap[b.Type]).Msg("getting ele center failed")
			b.Hidden = true
		}
	}

	b.Attributes, _ = ele.GetAttributes()
	b.NodeDepth = ele.Depth()
	b.InnerText = ele.GetInnerText()

	listeners, err := ele.GetEventListeners()
	// no listeners
	if err != nil || len(listeners) == 0 {
		//log.Debug().Str("type", browserk.HTMLTypeToStrMap[b.Type]).Msg("NO INTERACTABLES")
		return b
	}

	for _, listener := range listeners {
		eventType, ok := browserk.HTMLEventTypeMap[listener.Type]
		if !ok {
			eventType = browserk.HTMLEventcustom
		}
		line := strconv.Itoa(listener.LineNumber)
		col := strconv.Itoa(listener.ColumnNumber)

		b.Events[line+" "+col] = eventType
	}

	return b
}

func ElementToHTMLFormElement(ele *Element) *browserk.HTMLFormElement {
	b := &browserk.HTMLFormElement{Events: make(map[string]browserk.HTMLEventType, 0)}
	b.Attributes, _ = ele.GetAttributes()
	b.NodeDepth = ele.Depth()
	listeners, err := ele.GetEventListeners()
	if err == nil {
		for _, listener := range listeners {
			eventType, ok := browserk.HTMLEventTypeMap[listener.Type]
			if !ok {
				eventType = browserk.HTMLEventcustom
			}
			line := strconv.Itoa(listener.LineNumber)
			col := strconv.Itoa(listener.ColumnNumber)

			b.Events[line+" "+col] = eventType
		}
	}

	return b
}

// GCDRequestToBrowserk NetworkRequestWillBeSentEvent -> HTTPRequest
func GCDRequestToBrowserk(req *gcdapi.NetworkRequestWillBeSentEvent) *browserk.HTTPRequest {
	p := req.Params
	r := &browserk.HTTPRequest{
		RequestId:        p.RequestId,
		LoaderId:         p.LoaderId,
		DocumentURL:      p.DocumentURL,
		Request:          p.Request,
		Timestamp:        p.Timestamp,
		WallTime:         p.WallTime,
		Initiator:        p.Initiator,
		RedirectResponse: p.RedirectResponse,
		Type:             p.Type,
		FrameId:          p.FrameId,
		HasUserGesture:   p.HasUserGesture,
	}
	r.Hash()
	return r
}

// GCDResponseToBrowserk convert resp with body
// TODO: have a service check if we already have this body (via hash) and don't bother storing
func GCDResponseToBrowserk(resp *gcdapi.NetworkResponseReceivedEvent, body []byte) *browserk.HTTPResponse {
	p := resp.Params
	h := sha1.New()
	h.Write(body)

	r := &browserk.HTTPResponse{
		RequestId: p.RequestId,
		LoaderId:  p.LoaderId,
		Timestamp: p.Timestamp,
		Type:      p.Type,
		Response:  p.Response,
		FrameId:   p.FrameId,
		Body:      body,
		BodyHash:  h.Sum(nil),
	}
	r.Hash()
	return r
}

// GCDFetchRequestToIntercepted FetchRequestPausedEvent -> InterceptedHTTPRequest
func GCDFetchRequestToIntercepted(m *gcdapi.FetchRequestPausedEvent, container *Container) *browserk.InterceptedHTTPRequest {
	p := m.Params
	//r := container.GetRequest(p.RequestId)
	req := m.Params.Request
	//if r != nil { // <-- was this for redirects or something? i forgot.
	//	req = r.Request
	//}
	headers := make([]*gcdapi.FetchHeaderEntry, 0)
	if p.Request != nil && p.Request.Headers != nil {
		for k, v := range p.Request.Headers {
			switch rv := v.(type) {
			case string:
				headers = append(headers, &gcdapi.FetchHeaderEntry{Name: k, Value: rv})
			case []string:
				for _, value := range rv {
					headers = append(headers, &gcdapi.FetchHeaderEntry{Name: k, Value: value})
				}
			case nil:
				headers = append(headers, &gcdapi.FetchHeaderEntry{Name: k, Value: ""})
			default:
				log.Warn().Str("header_name", k).Msg("unable to encode header value")
			}
		}
	}

	i := &browserk.InterceptedHTTPRequest{
		RequestId:      p.RequestId,
		Request:        req,
		FrameId:        p.FrameId,
		ResourceType:   p.ResourceType,
		RequestHeaders: headers,
		NetworkId:      p.NetworkId,
		Modified: &browserk.HTTPModifiedRequest{
			RequestId: "",
			Url:       "",
			Method:    "",
			PostData:  "",
			Headers:   nil,
		},
	}
	i.Hash()
	return i
}

// GCDFetchResponseToIntercepted FetchRequestPausedEvent -> InterceptedHTTPResponse
func GCDFetchResponseToIntercepted(m *gcdapi.FetchRequestPausedEvent, body string, encoded bool) *browserk.InterceptedHTTPResponse {
	p := m.Params
	r := &browserk.InterceptedHTTPResponse{
		RequestId:           p.RequestId,
		Request:             p.Request,
		FrameId:             p.FrameId,
		ResourceType:        p.ResourceType,
		NetworkId:           p.NetworkId,
		ResponseErrorReason: p.ResponseErrorReason,
		ResponseHeaders:     p.ResponseHeaders,
		ResponseStatusCode:  p.ResponseStatusCode,
		Body:                body,
		BodyEncoded:         encoded,
		Modified: &browserk.HTTPModifiedResponse{
			ResponseCode:    0,
			ResponseHeaders: nil,
			Body:            nil,
			ResponsePhrase:  "",
		},
	}
	r.Hash()
	return r
}

// GCDCookieToBrowserk NetworkCookie -> Cookie
func GCDCookieToBrowserk(gcdCookie []*gcdapi.NetworkCookie) []*browserk.Cookie {
	if gcdCookie == nil {
		return nil
	}
	observed := time.Now()
	cookies := make([]*browserk.Cookie, len(gcdCookie))
	for i, c := range gcdCookie {
		cookies[i] = &browserk.Cookie{
			Name:         c.Name,
			Value:        c.Value,
			Domain:       c.Domain,
			Path:         c.Path,
			Expires:      c.Expires,
			Size:         c.Size,
			HTTPOnly:     c.HttpOnly,
			Secure:       c.Secure,
			Session:      c.Session,
			SameSite:     c.SameSite,
			Priority:     c.Priority,
			ObservedTime: observed,
		}
	}
	return cookies
}

// RedirectResponseToNetworkResponse NetworkRequestWillBeSentEvent (RedirectResponse) -> NetworkResponseReceivedEvent
func RedirectResponseToNetworkResponse(req *gcdapi.NetworkRequestWillBeSentEvent) *gcdapi.NetworkResponseReceivedEvent {
	p := req.Params
	return &gcdapi.NetworkResponseReceivedEvent{
		Method: "",
		Params: struct {
			RequestId string                  "json:\"requestId\""
			LoaderId  string                  "json:\"loaderId\""
			Timestamp float64                 "json:\"timestamp\""
			Type      string                  "json:\"type\""
			Response  *gcdapi.NetworkResponse "json:\"response\""
			FrameId   string                  "json:\"frameId,omitempty\""
		}{
			RequestId: p.RequestId,
			LoaderId:  p.LoaderId,
			Timestamp: p.RedirectResponse.Timing.ReceiveHeadersEnd,
			Type:      p.Type,
			Response:  p.RedirectResponse,
			FrameId:   p.FrameId,
		},
	}
}
