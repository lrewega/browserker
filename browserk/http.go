package browserk

import (
	"bytes"
	"crypto/md5"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack/v4"
	"github.com/wirepair/gcd/v2/gcdapi"
)

// HTTPRequest contains all information regarding a network request
type HTTPRequest struct {
	ID               []byte                   `json:"id"`
	RequestId        string                   `json:"requestId"`                  // Request identifier.
	LoaderId         string                   `json:"loaderId"`                   // Loader identifier. Empty string if the request is fetched from worker.
	DocumentURL      string                   `json:"documentURL"`                // URL of the document this request is loaded for.
	Request          *gcdapi.NetworkRequest   `json:"request"`                    // Request data from browser.
	Timestamp        float64                  `json:"timestamp"`                  // Timestamp.
	WallTime         float64                  `json:"wallTime"`                   // Timestamp.
	Initiator        *gcdapi.NetworkInitiator `json:"initiator"`                  // Request initiator.
	RedirectResponse *gcdapi.NetworkResponse  `json:"redirectResponse,omitempty"` // Redirect response data.
	Type             string                   `json:"type,omitempty"`             // Type of this resource. enum values: Document, Stylesheet, Image, Media, Font, Script, TextTrack, XHR, Fetch, EventSource, WebSocket, Manifest, SignedExchange, Ping, CSPViolationReport, Other
	FrameId          string                   `json:"frameId,omitempty"`          // Frame identifier.
	HasUserGesture   bool                     `json:"hasUserGesture,omitempty"`   // Whether the request is initiated by a user gesture. Defaults to false.
}

func (h *HTTPRequest) Hash() []byte {
	if h.ID != nil {
		return h.ID
	}
	hash := md5.New()
	hash.Write([]byte(h.Request.Method)) // TODO: make this better
	hash.Write(hashURL(h.Request.Url))
	hash.Write([]byte(h.Request.UrlFragment))
	hash.Write([]byte(h.Type))
	h.ID = hash.Sum(nil)
	return h.ID
}

func (h *HTTPRequest) StrHeaders() string {
	if h.Request == nil || h.Request.Headers == nil {
		return ""
	}
	headers := ""
	for n, v := range h.Request.Headers {
		switch t := v.(type) {
		case string:
			headers += n + ": " + t
		case []string:
			headers += n + ": " + strings.Join(t, ", ")
		}
	}
	return headers
}

// hashURL scheme, host (and port), path and sorted query names only (not values)
// TODO: Do analysis of query values to see if they are dynamic or static
// if static, they should be included: think x.jsp?page=admin&x=y vs x.jsp?page=user&x=y
func hashURL(in string) []byte {
	buf := &bytes.Buffer{}
	u, err := url.Parse(in)
	if err != nil {
		return []byte(in)
	}
	buf.Write([]byte(u.Scheme))
	buf.Write([]byte(u.Host))
	buf.Write([]byte(u.Path))
	queryNames := make([]string, 0)
	for k := range u.Query() {
		queryNames = append(queryNames, k)
	}
	sort.Strings(queryNames)
	buf.Write([]byte(strings.Join(queryNames, ",")))
	return buf.Bytes()
}

// Copy does a deep copy
// TODO: write a small astutil to generate deep copy with nil checks of nested objects
// for now, be super lazy
func (h *HTTPRequest) Copy() *HTTPRequest {
	if h == nil {
		return nil
	}
	d, err := msgpack.Marshal(h)
	if err != nil {
		panic("failed to copy HTTPRequest: " + err.Error())
	}
	c := &HTTPRequest{}
	if err = msgpack.Unmarshal(d, c); err != nil {
		panic("failed to copy HTTPRequest: " + err.Error())
	}
	return c
}

// HTTPResponse contains all information regarding a network response
type HTTPResponse struct {
	ID        []byte                  `json:"id"`
	RequestId string                  `json:"requestId"`           // Request identifier.
	LoaderId  string                  `json:"loaderId"`            // Loader identifier. Empty string if the request is fetched from worker.
	Timestamp float64                 `json:"timestamp"`           // Timestamp.
	Type      string                  `json:"type"`                // Resource type. enum values: Document, Stylesheet, Image, Media, Font, Script, TextTrack, XHR, Fetch, EventSource, WebSocket, Manifest, SignedExchange, Ping, CSPViolationReport, Other
	Response  *gcdapi.NetworkResponse `json:"response"`            // Response data.
	FrameId   string                  `json:"frameId,omitempty"`   // Frame identifier.
	Body      []byte                  `json:"body,omitempty"`      // Raw captured body data
	BodyHash  []byte                  `json:"body_hash,omitempty"` // sha1 hash of body data
}

func (h *HTTPResponse) Hash() []byte {
	if h.ID != nil {
		return h.ID
	}
	hash := md5.New()
	hash.Write([]byte(h.Response.MimeType)) // TODO: make this better
	hash.Write(hashURL(h.Response.Url))
	hash.Write(h.BodyHash)
	h.ID = hash.Sum(nil)
	return h.ID
}

func (h *HTTPResponse) StrHeaders() string {
	if h.Response == nil || h.Response.Headers == nil {
		return ""
	}
	headers := ""
	for n, v := range h.Response.Headers {
		switch t := v.(type) {
		case string:
			headers += n + ": " + t
		case []string:
			headers += n + ": " + strings.Join(t, ", ")
		}
	}
	return headers
}

// ResponseTimeMs returns how long the response took from when we finished sending
// to when we got response headers.
func (h *HTTPResponse) ResponseTimeMs() float64 {
	if h == nil || h.Response == nil || h.Response.Timing == nil {
		return 0
	}
	return h.Response.Timing.ReceiveHeadersEnd - h.Response.Timing.SendEnd
}

// Copy does a deep copy
// TODO: write a small astutil to generate deep copy with nil checks of nested objects
// for now, be super lazy
func (h *HTTPResponse) Copy() *HTTPResponse {
	if h == nil {
		return nil
	}
	d, err := msgpack.Marshal(h)
	if err != nil {
		panic("failed to copy HTTPResponse: " + err.Error())
	}
	c := &HTTPResponse{}
	if err = msgpack.Unmarshal(d, c); err != nil {
		panic("failed to copy HTTPResponse: " + err.Error())
	}
	return c
}

// InterceptedHTTPRequest contains all information regarding an intercepted request
type InterceptedHTTPRequest struct {
	ID             []byte                     `json:"id"`
	RequestId      string                     `json:"requestId"`                 // Each request the page makes will have a unique id.
	Request        *gcdapi.NetworkRequest     `json:"request"`                   // The details of the request.
	FrameId        string                     `json:"frameId"`                   // The id of the frame that initiated the request.
	ResourceType   string                     `json:"resourceType"`              // How the requested resource will be used. enum values: Document, Stylesheet, Image, Media, Font, Script, TextTrack, XHR, Fetch, EventSource, WebSocket, Manifest, SignedExchange, Ping, CSPViolationReport, Other
	RequestHeaders []*gcdapi.FetchHeaderEntry `json:"responseHeaders,omitempty"` // Response headers if intercepted at the response stage.
	NetworkId      string                     `json:"networkId,omitempty"`       // If the intercepted request had a corresponding Network.requestWillBeSent event fired for it, then this networkId will be the same as the requestId present in the requestWillBeSent event.
	SentTimestamp  time.Time                  `json:"sentTimestamp,omitempty"`
	Modified       *HTTPModifiedRequest
}

func (h *InterceptedHTTPRequest) Hash() []byte {
	if h.ID != nil {
		return h.ID
	}
	hash := md5.New()
	hash.Write([]byte(h.Request.Method)) // TODO: make this better
	hash.Write(hashURL(h.Request.Url))
	hash.Write([]byte(h.Request.UrlFragment))
	h.ID = hash.Sum(nil)
	return h.ID
}

// Copy does a deep copy
// TODO: write a small astutil to generate deep copy with nil checks of nested objects
// for now, be super lazy
func (h *InterceptedHTTPRequest) Copy() *InterceptedHTTPRequest {
	if h == nil {
		return nil
	}
	d, err := msgpack.Marshal(h)
	if err != nil {
		panic("failed to copy InterceptedHTTPRequest: " + err.Error())
	}
	c := &InterceptedHTTPRequest{}
	if err = msgpack.Unmarshal(d, c); err != nil {
		panic("failed to copy InterceptedHTTPRequest: " + err.Error())
	}
	return c
}

// HTTPModifiedRequest allow modifications
type HTTPModifiedRequest struct {
	ID        []byte                     `json:"id"`
	RequestId string                     `json:"requestId"`          // An id the client received in requestPaused event.
	Url       string                     `json:"url,omitempty"`      // If set, the request url will be modified in a way that's not observable by page.
	Method    string                     `json:"method,omitempty"`   // If set, the request method is overridden.
	PostData  string                     `json:"postData,omitempty"` // If set, overrides the post data in the request.
	Headers   []*gcdapi.FetchHeaderEntry `json:"headers,omitempty"`  // If set, overrides the request headers.
}

func (h *HTTPModifiedRequest) Hash() []byte {
	if h.ID != nil {
		return h.ID
	}
	hash := md5.New()
	hash.Write([]byte(h.Method)) // TODO: make this better
	hash.Write(hashURL(h.Url))
	hash.Write([]byte(h.PostData))
	h.ID = hash.Sum(nil)
	return h.ID
}

// Copy does a deep copy
// TODO: write a small astutil to generate deep copy with nil checks of nested objects
// for now, be super lazy
func (h *HTTPModifiedRequest) Copy() *HTTPModifiedRequest {
	if h == nil {
		return nil
	}
	d, err := msgpack.Marshal(h)
	if err != nil {
		panic("failed to copy HTTPModifiedRequest: " + err.Error())
	}
	c := &HTTPModifiedRequest{}
	if err = msgpack.Unmarshal(d, c); err != nil {
		panic("failed to copy HTTPModifiedRequest: " + err.Error())
	}
	return c
}

func (h *HTTPModifiedRequest) SetHeaders(headers map[string]interface{}) {
	if h.Headers == nil {
		h.Headers = make([]*gcdapi.FetchHeaderEntry, 0)
	}

	for k, value := range headers {
		switch v := value.(type) {
		case string:
			h.Headers = append(h.Headers, &gcdapi.FetchHeaderEntry{Name: k, Value: v})
		case []string:
			for _, header := range v {
				h.Headers = append(h.Headers, &gcdapi.FetchHeaderEntry{Name: k, Value: header})
			}
		default:
		}
	}
}

// InterceptedHTTPResponse to pass to middleware and allow modifications to Modified
type InterceptedHTTPResponse struct {
	ID                  []byte                     `json:"id"`
	RequestId           string                     `json:"requestId"`                     // Each request the page makes will have a unique id.
	Request             *gcdapi.NetworkRequest     `json:"request"`                       // The details of the request.
	FrameId             string                     `json:"frameId"`                       // The id of the frame that initiated the request.
	ResourceType        string                     `json:"resourceType"`                  // How the requested resource will be used. enum values: Document, Stylesheet, Image, Media, Font, Script, TextTrack, XHR, Fetch, EventSource, WebSocket, Manifest, SignedExchange, Ping, CSPViolationReport, Other
	ResponseErrorReason string                     `json:"responseErrorReason,omitempty"` // Response error if intercepted at response stage. enum values: Failed, Aborted, TimedOut, AccessDenied, ConnectionClosed, ConnectionReset, ConnectionRefused, ConnectionAborted, ConnectionFailed, NameNotResolved, InternetDisconnected, AddressUnreachable, BlockedByClient, BlockedByResponse
	ResponseStatusCode  int                        `json:"responseStatusCode,omitempty"`  // Response code if intercepted at response stage.
	ResponseHeaders     []*gcdapi.FetchHeaderEntry `json:"responseHeaders,omitempty"`     // Response headers if intercepted at the response stage.
	NetworkId           string                     `json:"networkId,omitempty"`           // If the intercepted request had a corresponding Network.requestWillBeSent event fired for it, then this networkId will be the same as the requestId present in the requestWillBeSent event.
	Body                string                     `json:"body,omitempty"`
	BodyEncoded         bool                       `json:"body_encoded,omitempty"`
	RecvTimestamp       time.Time                  `json:"recvTimestamp,omitempty"`
	Modified            *HTTPModifiedResponse
}

func (h *InterceptedHTTPResponse) Hash() []byte {
	if h.ID != nil {
		return h.ID
	}
	hash := md5.New()
	hash.Write([]byte(h.Request.Method)) // TODO: make this better
	hash.Write(hashURL(h.Request.Url))
	hash.Write([]byte(h.Request.UrlFragment))
	hash.Write([]byte(h.ResourceType))
	h.ID = hash.Sum(nil)
	return h.ID
}

// Copy does a deep copy
// TODO: write a small astutil to generate deep copy with nil checks of nested objects
// for now, be super lazy
func (h *InterceptedHTTPResponse) Copy() *InterceptedHTTPResponse {
	if h == nil {
		return nil
	}
	d, err := msgpack.Marshal(h)
	if err != nil {
		panic("failed to copy InterceptedHTTPResponse: " + err.Error())
	}
	c := &InterceptedHTTPResponse{}
	if err = msgpack.Unmarshal(d, c); err != nil {
		panic("failed to copy InterceptedHTTPResponse: " + err.Error())
	}
	return c
}

// HTTPModifiedResponse contains the modified http response data
type HTTPModifiedResponse struct {
	ID                    []byte                     `json:"id"`
	RequestId             string                     `json:"requestId"`
	ResponseCode          int                        `json:"responseCode"`                    // An HTTP response code.
	ResponseHeaders       []*gcdapi.FetchHeaderEntry `json:"responseHeaders,omitempty"`       // Response headers.
	BinaryResponseHeaders string                     `json:"binaryResponseHeaders,omitempty"` // Alternative way of specifying response headers as a \0-separated series of name: value pairs. Prefer the above method unless you need to represent some non-UTF8 values that can't be transmitted over the protocol as text.
	Body                  []byte                     `json:"body,omitempty"`                  // A response body.
	ResponsePhrase        string                     `json:"responsePhrase,omitempty"`        // A textual representation of responseCode. If absent, a standard phrase matching responseCode is used.
}

func (h *HTTPModifiedResponse) Hash() []byte {
	if h.ID != nil {
		return h.ID
	}
	hash := md5.New()
	hash.Write([]byte(h.Body)) // TODO: make this better
	hash.Write([]byte(h.ResponsePhrase))
	h.ID = hash.Sum(nil)
	return h.ID
}

// Copy does a deep copy
// TODO: write a small astutil to generate deep copy with nil checks of nested objects
// for now, be super lazy
func (h *HTTPModifiedResponse) Copy() *HTTPModifiedResponse {
	if h == nil {
		return nil
	}
	d, err := msgpack.Marshal(h)
	if err != nil {
		panic("failed to copy HTTPModifiedResponse: " + err.Error())
	}
	c := &HTTPModifiedResponse{}
	if err = msgpack.Unmarshal(d, c); err != nil {
		panic("failed to copy HTTPModifiedResponse: " + err.Error())
	}
	return c
}
