package browser

import (
	"crypto/sha1"

	"github.com/pkg/errors"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
	"gitlab.com/browserker/browserk"
)

// https://chromium.googlesource.com/chromium/src/+/master/third_party/WebKit/Source/core/inspector/InspectorNetworkAgent.cpp#96
const maximumTotalBufferSize = -1

const maximumResourceBufferSize = -1

const maximumPostDataSize = -1

// GcdResponseFunc internal response function type
type GcdResponseFunc func(target *gcd.ChromeTarget, payload []byte)

// TabDisconnectedHandler is called when the tab crashes or the inspector was disconnected
type TabDisconnectedHandler func(tab *Tab, reason string)

// PromptHandlerFunc function to handle javascript dialog prompts as they occur, pass to SetJavaScriptPromptHandler
// Internally this should call tab.Page.HandleJavaScriptDialog(accept bool, promptText string)
type PromptHandlerFunc func(tab *Tab, message, promptType string)

// ConsoleMessageFunc function for handling console messages
type ConsoleMessageFunc func(tab *Tab, message *gcdapi.ConsoleConsoleMessage)

// NetworkRequestHandlerFunc function for handling network requests
type NetworkRequestHandlerFunc func(tab *Tab, request *NetworkRequest)

// NetworkResponseHandlerFunc function for handling network responses
type NetworkResponseHandlerFunc func(tab *Tab, response *NetworkResponse)

// NetworkFinishedHandlerFunc function for handling network finished, meaning it's safe to call Network.GetResponseBody
type NetworkFinishedHandlerFunc func(tab *Tab, requestID string, dataLength, timeStamp float64)

// StorageFunc function for ListenStorageEvents returns the eventType of cleared, updated, removed or added.
type StorageFunc func(tab *Tab, eventType string, eventDetails *StorageEvent)

// DomChangeHandlerFunc function to listen for DOM Node Change Events
type DomChangeHandlerFunc func(tab *Tab, change *NodeChangeEvent)

// ConditionalFunc function to iteratively call until returns without error
type ConditionalFunc func(tab *Tab) bool

// revive:exported
var (
	ErrNavigationTimedOut = errors.New("navigation timed out")
	ErrTabCrashed         = errors.New("tab crashed")
	ErrTabClosing         = errors.New("closing")
	ErrTimedOut           = errors.New("request timed out")
	ErrNavigating         = errors.New("error in navigation")
	ErrBrowserClosing     = errors.New("unable to load, as closing down")
)

// ElementNotFoundErr when we are unable to find an element/nodeID
type ElementNotFoundErr struct {
	Message string
}

func (e *ElementNotFoundErr) Error() string {
	return "Unable to find element " + e.Message
}

// InvalidTabErr when we are unable to access a tab
type InvalidTabErr struct {
	Message string
}

func (e *InvalidTabErr) Error() string {
	return "Unable to access tab: " + e.Message
}

// InvalidNavigationErr when unable to navigate Forward or Back
type InvalidNavigationErr struct {
	Message string
}

func (e *InvalidNavigationErr) Error() string {
	return e.Message
}

// ScriptEvaluationErr returned when an injected script caused an error
type ScriptEvaluationErr struct {
	Message          string
	ExceptionText    string
	ExceptionDetails *gcdapi.RuntimeExceptionDetails
}

func (e *ScriptEvaluationErr) Error() string {
	return e.Message + " " + e.ExceptionText
}

// TimeoutErr when Tab.Navigate has timed out
type TimeoutErr struct {
	Message string
}

func (e *TimeoutErr) Error() string {
	return "Timed out " + e.Message
}

// NodeType are standard browser node types
type NodeType uint8

// revive:exported
const (
	NodeElement               NodeType = 0x1
	NodeText                  NodeType = 0x3
	NodeProcessingInstruction NodeType = 0x7
	NodeComment               NodeType = 0x8
	NodeDocument              NodeType = 0x9
	NodeDocumentType          NodeType = 0x10
	NodeDocumentFragment      NodeType = 0x11
)

var nodeTypeMap = map[NodeType]string{
	NodeElement:               "ELEMENT_NODE",
	NodeText:                  "TEXT_NODE",
	NodeProcessingInstruction: "PROCESSING_INSTRUCTION_NODE",
	NodeComment:               "COMMENT_NODE",
	NodeDocument:              "DOCUMENT_NODE",
	NodeDocumentType:          "DOCUMENT_TYPE_NODE",
	NodeDocumentFragment:      "DOCUMENT_FRAGMENT_NODE",
}

// ChangeEventType Document/Node change event types
type ChangeEventType uint16

// revive:exported
const (
	DocumentUpdatedEvent        ChangeEventType = 0x0
	SetChildNodesEvent          ChangeEventType = 0x1
	AttributeModifiedEvent      ChangeEventType = 0x2
	AttributeRemovedEvent       ChangeEventType = 0x3
	InlineStyleInvalidatedEvent ChangeEventType = 0x4
	CharacterDataModifiedEvent  ChangeEventType = 0x5
	ChildNodeCountUpdatedEvent  ChangeEventType = 0x6
	ChildNodeInsertedEvent      ChangeEventType = 0x7
	ChildNodeRemovedEvent       ChangeEventType = 0x8
)

var changeEventMap = map[ChangeEventType]string{
	DocumentUpdatedEvent:        "DocumentUpdatedEvent",
	SetChildNodesEvent:          "SetChildNodesEvent",
	AttributeModifiedEvent:      "AttributeModifiedEvent",
	AttributeRemovedEvent:       "AttributeRemovedEvent",
	InlineStyleInvalidatedEvent: "InlineStyleInvalidatedEvent",
	CharacterDataModifiedEvent:  "CharacterDataModifiedEvent",
	ChildNodeCountUpdatedEvent:  "ChildNodeCountUpdatedEvent",
	ChildNodeInsertedEvent:      "ChildNodeInsertedEvent",
	ChildNodeRemovedEvent:       "ChildNodeRemovedEvent",
}

func (evt ChangeEventType) String() string {
	if s, ok := changeEventMap[evt]; ok {
		return s
	}
	return ""
}

// NodeChangeEvent for handling DOM updating nodes
type NodeChangeEvent struct {
	EventType      ChangeEventType   // the type of node change event
	NodeID         int               // nodeid of change
	NodeIDs        []int             // nodeid of changes for inlinestyleinvalidated
	ChildNodeCount int               // updated childnodecount event
	Nodes          []*gcdapi.DOMNode // Child nodes array. for setChildNodesEvent
	Node           *gcdapi.DOMNode   // node for child node inserted event
	Name           string            // attribute name
	Value          string            // attribute value
	CharacterData  string            // new text value for characterDataModified events
	ParentNodeID   int               // node id for setChildNodesEvent, childNodeInsertedEvent and childNodeRemovedEvent
	PreviousNodeID int               // previous node id for childNodeInsertedEvent
}

// StorageEventType of storage related events.
type StorageEventType uint16

// StorageEvent details
type StorageEvent struct {
	IsLocalStorage bool   // if true, local storage, false session storage
	SecurityOrigin string // origin that this event occurred on
	Key            string // storage key
	NewValue       string // new storage value
	OldValue       string // old storage value
}

func GCDRequestToBrowserk(req *gcdapi.NetworkRequestWillBeSentEvent) *browserk.HTTPRequest {
	p := req.Params
	return &browserk.HTTPRequest{
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
}

// GCDResponseToBrowserk convert resp with body
// TODO: have a service check if we already have this body (via hash) and don't bother storing
func GCDResponseToBrowserk(resp *gcdapi.NetworkResponseReceivedEvent, body []byte) *browserk.HTTPResponse {
	p := resp.Params
	h := sha1.New()
	h.Write(body)

	return &browserk.HTTPResponse{
		RequestId: p.RequestId,
		LoaderId:  p.LoaderId,
		Timestamp: p.Timestamp,
		Type:      p.Type,
		Response:  p.Response,
		FrameId:   p.FrameId,
		Body:      body,
		BodyHash:  h.Sum(nil),
	}
}
