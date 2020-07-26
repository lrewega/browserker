package browserk

import "context"

// InjectionType determines where we are injectin (name/value/index)
type InjectionType string

const (
	InjectValue InjectionType = "value" // the default for values/paths etc
	InjectName  InjectionType = "name"  // for parameter/query names
	InjectIndex InjectionType = "index" // for query[index] values
)

type InjectionPos int

// All node types implement the InjectionNode interface.
type InjectionNode interface {
	Pos() InjectionPos // position of first character belonging to the node
	End() InjectionPos // position of first character immediately after the node
}

type InjectionExpr interface {
	InjectionNode
	String() string
	Inject(string, InjectionType) bool
	Reset()
	Loc() InjectionLocation
	//Next() InjectionExpr
}

// InjectionLocation for configuring plugins where they will attack
type InjectionLocation int

// Has determines if this injection location contains the loc
func (i InjectionLocation) Has(loc InjectionLocation) bool {
	return i&loc != 0
}

func (i InjectionLocation) HasIn(locs []InjectionLocation) bool {
	for _, loc := range locs {
		if i.Has(loc) {
			return true
		}
	}
	return false
}

// Injection Location points
const (
	_            InjectionLocation = iota
	InjectMethod InjectionLocation = 1 << iota
	InjectPath
	InjectFile
	InjectQuery
	InjectQueryName
	InjectQueryValue
	InjectQueryIndex
	InjectFragment
	InjectHeader
	InjectHeaderName
	InjectHeaderValue
	InjectCookie
	InjectCookieName
	InjectCookieValue
	InjectBody
	InjectBodyName
	InjectBodyValue
	InjectBodyIndex
	InjectJSON
	InjectJSONName
	InjectJSONValue
	InjectXML
	InjectXMLName
	InjectXMLValue
)

const (
	// InjectAll injects into literally any point we can
	InjectAll InjectionLocation = InjectMethod | InjectPath | InjectFile | InjectQuery | InjectQueryName | InjectQueryValue | InjectQueryIndex | InjectFragment | InjectHeader | InjectHeaderName | InjectHeaderValue | InjectCookie | InjectCookieName | InjectCookieValue | InjectBody | InjectBodyName | InjectBodyValue | InjectBodyIndex | InjectJSON | InjectJSONName | InjectJSONValue | InjectXML | InjectXMLName | InjectXMLValue
	// InjectCommon injects into common path/value parameters
	InjectCommon InjectionLocation = InjectPath | InjectFile | InjectQuery | InjectQueryValue | InjectFragment | InjectHeaderValue | InjectCookieValue | InjectBody | InjectBodyValue | InjectJSON | InjectJSONValue | InjectXML | InjectXMLValue
	// InjectNameValue Names and Values
	InjectNameValue InjectionLocation = InjectQuery | InjectQueryValue | InjectQueryName | InjectQueryIndex | InjectHeaderName | InjectHeaderValue | InjectCookieValue | InjectBody | InjectBodyName | InjectBodyValue | InjectJSON | InjectJSONName | InjectJSONValue | InjectXML | InjectXMLName | InjectXMLValue
	// InjectValues only TODO doesn't work since we don't peek into KV yet
	InjectValues InjectionLocation = InjectQuery | InjectQueryValue | InjectHeaderValue | InjectJSONValue | InjectBody | InjectBodyValue | InjectXMLValue
)

func (i InjectionLocation) String() string {
	switch i {
	case InjectMethod:
		return "InjectMethod"
	case InjectPath:
		return "InjectPath"
	case InjectFile:
		return "InjectFile"
	case InjectQuery:
		return "InjectQuery"
	case InjectQueryName:
		return "InjectQueryName"
	case InjectQueryValue:
		return "InjectQueryValue"
	case InjectQueryIndex:
		return "InjectQueryIndex"
	case InjectFragment:
		return "InjectFragment"
	case InjectHeader:
		return "InjectHeader"
	case InjectHeaderName:
		return "InjectHeaderName"
	case InjectHeaderValue:
		return "InjectHeaderValue"
	case InjectCookie:
		return "InjectCookie"
	case InjectCookieName:
		return "InjectCookieName"
	case InjectCookieValue:
		return "InjectCookieValue"
	case InjectBody:
		return "InjectBody"
	case InjectBodyName:
		return "InjectBodyName"
	case InjectBodyValue:
		return "InjectBodyValue"
	case InjectBodyIndex:
		return "InjectBodyIndex"
	case InjectJSON:
		return "InjectJSON"
	case InjectJSONName:
		return "InjectJSONName"
	case InjectJSONValue:
		return "InjectJSONValue"
	case InjectXML:
		return "InjectXML"
	case InjectXMLName:
		return "InjectXMLName"
	case InjectXMLValue:
		return "InjectXMLValue"
	}
	return ""
}

// Injector handles injecting into target requests/pages using different methods
type Injector interface {
	Browser() Browser
	BCtx() *Context
	Message() *HTTPMessage
	InjectionExpr() InjectionExpr
	Nav() *Navigation
	NavResultID() []byte
	ReplacePath(newValue string, index int)
	ReplaceFile(newValue string)
	ReplaceURI(newURI string)
	ReplaceHeader(name, value string)
	AddHeader(name, value string)
	RemoveHeader(name string)
	ReplaceBody(newBody []byte)
	Send(ctx context.Context, withRender bool) (*InterceptedHTTPMessage, error)
	SendNew(ctx context.Context, req *HTTPRequest, withRender bool) (*InterceptedHTTPMessage, error)
	// BrowserSend ..? (for xss/plugins that send through the current page).
}
