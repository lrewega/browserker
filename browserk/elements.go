package browserk

import (
	"bytes"
	"crypto/md5"
	"sort"
	"strings"
)

type ActHTMLElement interface {
	Tag() string
	ElementType() HTMLElementType
	GetAttribute(name string) string
	AllAttributes() map[string]string
	Depth() int
	Hash() []byte
}

// HTMLElement type
type HTMLElement struct {
	Type          HTMLElementType
	CustomTagName string
	Events        map[string]HTMLEventType // key: line,col => event type
	Attributes    map[string]string
	InnerText     string
	Hidden        bool
	NodeDepth     int
	ID            []byte
	Value         string // value to set if it's an input field or whatever
}

// Hash the element to (hopefully) a unique value
// Don't include depth as it may change but we can check that individually
// for optimization purposes if Depth && ID == ...
func (h *HTMLElement) Hash() []byte {
	if h.ID != nil {
		return h.ID
	}
	hash := md5.New()

	if h.Type == HASHTEXT {
		hash.Write([]byte(h.InnerText))
		h.ID = hash.Sum(nil)
		return h.ID
	}

	vals := ImportantAttributeValues(h.Type, h.Attributes)
	vals = append(vals, h.InnerText)
	sort.StringSlice(vals).Sort()
	sorted := strings.Join(vals, "")
	hash.Write([]byte(sorted))
	// include event line/col into the uniqueness
	evts := sortEvents(h.Events)
	hash.Write([]byte(evts))
	if h.Hidden {
		hash.Write([]byte{1})
	} else {
		hash.Write([]byte{0})
	}
	h.ID = hash.Sum(nil)
	return h.ID
}

func (h *HTMLElement) ElementType() HTMLElementType {
	return h.Type
}

func (h *HTMLElement) Tag() string {
	tag := strings.ToLower(HTMLTypeToStrMap[h.Type])
	if h.Type == CUSTOM {
		tag = h.CustomTagName
	}
	return tag
}

func (h *HTMLElement) Depth() int {
	return h.NodeDepth
}

func (h *HTMLElement) GetAttribute(name string) string {
	val, exist := h.Attributes[name]
	if !exist {
		return ""
	}
	return val
}

func (h *HTMLElement) AllAttributes() map[string]string {
	return h.Attributes
}

// FormType determine what type of form it is
type FormType int8

//revive:disable:var-export
const (
	FormLogin FormType = iota
	FormLogout
	FormUserRegistration
	FormSignUp // letters etc
	FormContact
	FormComment
	FormSearch
	FormCreate
	FormDelete
	FormEdit
	FormBilling
	FormAddress
	// TODO: Add more, or scan the top 1mil sites and extract all forms or something
)

// HTMLFormElement and it's children
type HTMLFormElement struct {
	Type           HTMLElementType
	FormType       FormType
	Events         map[string]HTMLEventType
	Attributes     map[string]string
	Hidden         bool
	NodeDepth      int
	ChildElements  []*HTMLElement // capture all children (labels etc) so we can do context analysis
	ID             []byte
	SubmitButtonID []byte
}

// Hash the form and it's input elements to (hopefully) a unique value
// Don't include depth as it may change but we can check that individually
func (h *HTMLFormElement) Hash() []byte {
	if h.ID != nil {
		return h.ID
	}
	hash := md5.New()

	vals := ImportantAttributeValues(FORM, h.Attributes)
	/*
		adds too much variabliity possibly...
		for _, child := range h.ChildElements {
			if child.Type != INPUT {
				continue
			}
			vals = append(vals, ImportantAttributeValues(INPUT, child.Attributes)...)
		}
	*/
	sort.StringSlice(vals).Sort()
	sorted := strings.Join(vals, "")
	hash.Write([]byte(sorted))
	evts := sortEvents(h.Events)
	hash.Write([]byte(evts))
	if h.Hidden {
		hash.Write([]byte{1})
	} else {
		hash.Write([]byte{0})
	}
	h.ID = hash.Sum(nil)
	return h.ID
}

func (h *HTMLFormElement) GetAttribute(name string) string {
	val, exist := h.Attributes[name]
	if !exist {
		return ""
	}
	return val
}

func (h *HTMLFormElement) ElementType() HTMLElementType {
	return h.Type
}

func (h *HTMLFormElement) AllAttributes() map[string]string {
	return h.Attributes
}

func (h *HTMLFormElement) Tag() string {
	return HTMLTypeToStrMap[h.Type]
}

func (h *HTMLFormElement) Depth() int {
	return h.NodeDepth
}

// GetNextOf returns a child element of the requested type if it exists
func (h *HTMLFormElement) GetNextOf(start int, elementType HTMLElementType) *HTMLElement {
	if start > len(h.ChildElements)-1 {
		return nil
	}

	for i := start; i < len(h.ChildElements); i++ {
		if h.ChildElements[i].Type == elementType {
			return h.ChildElements[i]
		}
	}
	return nil
}

func (h *HTMLFormElement) GetChildByNameOrID(name string) *HTMLElement {
	for _, ele := range h.ChildElements {
		if ele.GetAttribute("name") == name || ele.GetAttribute("id") == name {
			return ele
		}
	}
	return nil
}

func (h *HTMLFormElement) GetChildByHash(hash []byte) *HTMLElement {
	for _, ele := range h.ChildElements {
		if bytes.Compare(hash, ele.Hash()) == 0 {
			return ele
		}
	}
	return nil
}

func sortEvents(toSort map[string]HTMLEventType) string {
	events := make([]string, len(toSort))
	i := 0
	for k := range toSort {
		events[i] = k
		i++
	}
	sort.StringSlice(events).Sort()
	sorted := strings.Join(events, "")
	return sorted
}

// ImportantAttributeValues extracts the values for important attributes depending on HTMLElementType
func ImportantAttributeValues(elementType HTMLElementType, attrs map[string]string) []string {
	vals := make([]string, 0)
	// TODO: Add more/all
	for k, v := range attrs {
		switch elementType {
		case LINK:
			switch k {
			case "rel", "title":
				vals = append(vals, v)
			}
		case FORM:
			switch k {
			case "action", "method", "accept-charset", "autocomplete", "enctype", "target":
				vals = append(vals, v)
			}
		case INPUT:
			switch k {
			case "placeholder", "aria-label", "type", "name", "id":
				vals = append(vals, v)
			}
		case LABEL:
			switch k {
			case "for":
				vals = append(vals, v)
			}
		case META:
			switch k {
			case "property", "content":
				vals = append(vals, v)
			}
		case A:
			switch k {
			case "href":
				vals = append(vals, v)
			}
		case IMG, IFRAME, FRAME, SCRIPT, EMBED, OBJECT:
			switch k {
			case "src":
				vals = append(vals, v)
			}
		}
	}
	// always add class if it exists need to becareful for state switches (class = disabled etc)
	// and also potential for dynamic fields (maybe test if dynamic before adding?)
	// TODO: strip out disabled/other keywords
	if class, ok := attrs["class"]; ok {
		vals = append(vals, class)
	}

	return vals
}

// HTMLElementType tag name
type HTMLElementType int16

// revive:disable:var-naming
const (
	INVALID HTMLElementType = iota
	// Main Root
	HTML

	// METADATA
	BASE
	HEAD
	LINK
	META
	STYLE
	TITLE

	// Sectioning Root
	BODY

	//Content
	ADDRESS
	ARTICLE
	ASIDE
	FOOTER
	HEADER
	H1
	H2
	H3
	H4
	H5
	H6
	HGROUP
	MAIN
	NAV
	SECTION

	// Text
	BLOCKQUOTE
	DD
	DIV
	DL
	DT
	FIGCAPTION
	FIGURE
	HR
	LI
	// MAIN
	OL
	P
	PRE
	UL

	// Inline Text
	A
	ABBR
	B
	BDI
	BDO
	BR
	CITE
	CODE
	DATA
	DFN
	EM
	I
	KBD
	MARK
	Q
	RB
	RP
	RT
	RTC
	RUBY
	S
	SAMP
	SMALL
	SPAN
	STRONG
	SUB
	SUP
	TIME
	U
	VAR
	WBR

	// Image and Multimedia
	AREA
	AUDIO
	IMG
	MAP
	TRACK
	VIDEO

	// Embedded
	EMBED
	IFRAME
	OBJECT
	PARAM
	PICTURE
	SOURCE

	// Scripting
	CANVAS
	NOSCRIPT
	SCRIPT

	// Demarcating
	DEL
	INS

	// Table Content
	CAPTION
	COL
	COLGROUP
	TABLE
	TBODY
	TD
	TFOOT
	TH
	THEAD
	TR

	// Forms
	BUTTON
	DATALIST
	FIELDSET
	FORM
	INPUT
	LABEL
	LEGEND
	METER
	OPTGROUP
	OPTION
	OUTPUT
	PROGRESS
	SELECT
	TEXTAREA

	// Interactive
	DETAILS
	DIALOG
	MENU
	SUMMARY

	// Web Components
	SLOT
	TEMPLATE

	// Deprecated
	ACRONYM
	APPLET
	BASEFONT
	BGSOUND
	BIG
	BLINK
	CENTER
	COMMAND
	CONTENT
	DIR
	ELEMENT
	FONT
	FRAME
	FRAMESET
	IMAGE
	ISINDEX
	KEYGEN
	LISTING
	MARQUEE
	MENUITEM
	MULTICOL
	NEXTID
	NOBR
	NOEMBED
	NOFRAMES
	PLAINTEXT
	SHADOW
	SPACER
	STRIKE
	TT
	XMP

	// Custom/Non-standard
	CUSTOM
	HASHTEXT
)

// HTMLTypeMap for taking in tag name -> outputing HTMLElementType
var HTMLTypeMap = map[string]HTMLElementType{
	"HTML": HTML,
	// METADATA
	"BASE":  BASE,
	"HEAD":  HEAD,
	"LINK":  LINK,
	"META":  META,
	"STYLE": STYLE,
	"TITLE": TITLE,

	// Sectioning Root
	"BODY": BODY,

	//Content
	"ADDRESS": ADDRESS,
	"ARTICLE": ARTICLE,
	"ASIDE":   ASIDE,
	"FOOTER":  FOOTER,
	"HEADER":  HEADER,
	"H1":      H1,
	"H2":      H2,
	"H3":      H3,
	"H4":      H4,
	"H5":      H5,
	"H6":      H6,
	"HGROUP":  HGROUP,
	"MAIN":    MAIN,
	"NAV":     NAV,
	"SECTION": SECTION,

	// Text
	"BLOCKQUOTE": BLOCKQUOTE,
	"DD":         DD,
	"DIV":        DIV,
	"DL":         DL,
	"DT":         DT,
	"FIGCAPTION": FIGCAPTION,
	"FIGURE":     FIGURE,
	"HR":         HR,
	"LI":         LI,
	// MAIN
	"OL":  OL,
	"P":   P,
	"PRE": PRE,
	"UL":  UL,

	// Inline Text
	"A":      A,
	"ABBR":   ABBR,
	"B":      B,
	"BDI":    BDI,
	"BDO":    BDO,
	"BR":     BR,
	"CITE":   CITE,
	"CODE":   CODE,
	"DATA":   DATA,
	"DFN":    DFN,
	"EM":     EM,
	"I":      I,
	"KBD":    KBD,
	"MARK":   MARK,
	"Q":      Q,
	"RB":     RB,
	"RP":     RP,
	"RT":     RT,
	"RTC":    RTC,
	"RUBY":   RUBY,
	"S":      S,
	"SAMP":   SAMP,
	"SMALL":  SMALL,
	"SPAN":   SPAN,
	"STRONG": STRONG,
	"SUB":    SUB,
	"SUP":    SUP,
	"TIME":   TIME,
	"U":      U,
	"VAR":    VAR,
	"WBR":    WBR,

	// Image and Multimedia
	"AREA":  AREA,
	"AUDIO": AUDIO,
	"IMG":   IMG,
	"MAP":   MAP,
	"TRACK": TRACK,
	"VIDEO": VIDEO,

	// Embedded
	"EMBED":   EMBED,
	"IFRAME":  IFRAME,
	"OBJECT":  OBJECT,
	"PARAM":   PARAM,
	"PICTURE": PICTURE,
	"SOURCE":  SOURCE,

	// Scripting
	"CANVAS":   CANVAS,
	"NOSCRIPT": NOSCRIPT,
	"SCRIPT":   SCRIPT,

	// Demarcating
	"DEL": DEL,
	"INS": INS,

	// Table Content
	"CAPTION":  CAPTION,
	"COL":      COL,
	"COLGROUP": COLGROUP,
	"TABLE":    TABLE,
	"TBODY":    TBODY,
	"TD":       TD,
	"TFOOT":    TFOOT,
	"TH":       TH,
	"THEAD":    THEAD,
	"TR":       TR,

	// Forms
	"BUTTON":   BUTTON,
	"DATALIST": DATALIST,
	"FIELDSET": FIELDSET,
	"FORM":     FORM,
	"INPUT":    INPUT,
	"LABEL":    LABEL,
	"LEGEND":   LEGEND,
	"METER":    METER,
	"OPTGROUP": OPTGROUP,
	"OPTION":   OPTION,
	"OUTPUT":   OUTPUT,
	"PROGRESS": PROGRESS,
	"SELECT":   SELECT,
	"TEXTAREA": TEXTAREA,

	// Interactive
	"DETAILS": DETAILS,
	"DIALOG":  DIALOG,
	"MENU":    MENU,
	"SUMMARY": SUMMARY,

	// Web Components
	"SLOT":     SLOT,
	"TEMPLATE": TEMPLATE,

	// Deprecated
	"ACRONYM":   ACRONYM,
	"APPLET":    APPLET,
	"BASEFONT":  BASEFONT,
	"BGSOUND":   BGSOUND,
	"BIG":       BIG,
	"BLINK":     BLINK,
	"CENTER":    CENTER,
	"COMMAND":   COMMAND,
	"CONTENT":   CONTENT,
	"DIR":       DIR,
	"ELEMENT":   ELEMENT,
	"FONT":      FONT,
	"FRAME":     FRAME,
	"FRAMESET":  FRAMESET,
	"IMAGE":     IMAGE,
	"ISINDEX":   ISINDEX,
	"KEYGEN":    KEYGEN,
	"LISTING":   LISTING,
	"MARQUEE":   MARQUEE,
	"MENUITEM":  MENUITEM,
	"MULTICOL":  MULTICOL,
	"NEXTID":    NEXTID,
	"NOBR":      NOBR,
	"NOEMBED":   NOEMBED,
	"NOFRAMES":  NOFRAMES,
	"PLAINTEXT": PLAINTEXT,
	"SHADOW":    SHADOW,
	"SPACER":    SPACER,
	"STRIKE":    STRIKE,
	"TT":        TT,
	"XMP":       XMP,
}

var HTMLTypeToStrMap = map[HTMLElementType]string{
	HTML: "HTML",
	// METADATA
	BASE:  "BASE",
	HEAD:  "HEAD",
	LINK:  "LINK",
	META:  "META",
	STYLE: "STYLE",
	TITLE: "TITLE",

	// Sectioning Root
	BODY: "BODY",

	//Content
	ADDRESS: "ADDRESS",
	ARTICLE: "ARTICLE",
	ASIDE:   "ASIDE",
	FOOTER:  "FOOTER",
	HEADER:  "HEADER",
	H1:      "H1",
	H2:      "H2",
	H3:      "H3",
	H4:      "H4",
	H5:      "H5",
	H6:      "H6",
	HGROUP:  "HGROUP",
	MAIN:    "MAIN",
	NAV:     "NAV",
	SECTION: "SECTION",

	// Text
	BLOCKQUOTE: "BLOCKQUOTE",
	DD:         "DD",
	DIV:        "DIV",
	DL:         "DL",
	DT:         "DT",
	FIGCAPTION: "FIGCAPTION",
	FIGURE:     "FIGURE",
	HR:         "HR",
	LI:         "LI",
	// MAIN
	OL:  "OL",
	P:   "P",
	PRE: "PRE",
	UL:  "UL",

	// Inline Text
	A:      "A",
	ABBR:   "ABBR",
	B:      "B",
	BDI:    "BDI",
	BDO:    "BDO",
	BR:     "BR",
	CITE:   "CITE",
	CODE:   "CODE",
	DATA:   "DATA",
	DFN:    "DFN",
	EM:     "EM",
	I:      "I",
	KBD:    "KBD",
	MARK:   "MARK",
	Q:      "Q",
	RB:     "RB",
	RP:     "RP",
	RT:     "RT",
	RTC:    "RTC",
	RUBY:   "RUBY",
	S:      "S",
	SAMP:   "SAMP",
	SMALL:  "SMALL",
	SPAN:   "SPAN",
	STRONG: "STRONG",
	SUB:    "SUB",
	SUP:    "SUP",
	TIME:   "TIME",
	U:      "U",
	VAR:    "VAR",
	WBR:    "WBR",

	// Image and Multimedia
	AREA:  "AREA",
	AUDIO: "AUDIO",
	IMG:   "IMG",
	MAP:   "MAP",
	TRACK: "TRACK",
	VIDEO: "VIDEO",

	// Embedded
	EMBED:   "EMBED",
	IFRAME:  "IFRAME",
	OBJECT:  "OBJECT",
	PARAM:   "PARAM",
	PICTURE: "PICTURE",
	SOURCE:  "SOURCE",

	// Scripting
	CANVAS:   "CANVAS",
	NOSCRIPT: "NOSCRIPT",
	SCRIPT:   "SCRIPT",

	// Demarcating
	DEL: "DEL",
	INS: "INS",

	// Table Content
	CAPTION:  "CAPTION",
	COL:      "COL",
	COLGROUP: "COLGROUP",
	TABLE:    "TABLE",
	TBODY:    "TBODY",
	TD:       "TD",
	TFOOT:    "TFOOT",
	TH:       "TH",
	THEAD:    "THEAD",
	TR:       "TR",

	// Forms
	BUTTON:   "BUTTON",
	DATALIST: "DATALIST",
	FIELDSET: "FIELDSET",
	FORM:     "FORM",
	INPUT:    "INPUT",
	LABEL:    "LABEL",
	LEGEND:   "LEGEND",
	METER:    "METER",
	OPTGROUP: "OPTGROUP",
	OPTION:   "OPTION",
	OUTPUT:   "OUTPUT",
	PROGRESS: "PROGRESS",
	SELECT:   "SELECT",
	TEXTAREA: "TEXTAREA",

	// Interactive
	DETAILS: "DETAILS",
	DIALOG:  "DIALOG",
	MENU:    "MENU",
	SUMMARY: "SUMMARY",

	// Web Components
	SLOT:     "SLOT",
	TEMPLATE: "TEMPLATE",

	// Deprecated
	ACRONYM:   "ACRONYM",
	APPLET:    "APPLET",
	BASEFONT:  "BASEFONT",
	BGSOUND:   "BGSOUND",
	BIG:       "BIG",
	BLINK:     "BLINK",
	CENTER:    "CENTER",
	COMMAND:   "COMMAND",
	CONTENT:   "CONTENT",
	DIR:       "DIR",
	ELEMENT:   "ELEMENT",
	FONT:      "FONT",
	FRAME:     "FRAME",
	FRAMESET:  "FRAMESET",
	IMAGE:     "IMAGE",
	ISINDEX:   "ISINDEX",
	KEYGEN:    "KEYGEN",
	LISTING:   "LISTING",
	MARQUEE:   "MARQUEE",
	MENUITEM:  "MENUITEM",
	MULTICOL:  "MULTICOL",
	NEXTID:    "NEXTID",
	NOBR:      "NOBR",
	NOEMBED:   "NOEMBED",
	NOFRAMES:  "NOFRAMES",
	PLAINTEXT: "PLAINTEXT",
	SHADOW:    "SHADOW",
	SPACER:    "SPACER",
	STRIKE:    "STRIKE",
	TT:        "TT",
	XMP:       "XMP",
}

// HTMLEventType for various HTML events
type HTMLEventType int8

// revive:exported
const (
	// Resource events
	HTMLEventerror        HTMLEventType = iota //  A resource failed to load.
	HTMLEventabort                             //  The loading of a resource has been aborted.
	HTMLEventload                              //  A resource and its dependent resources have finished loading.
	HTMLEventbeforeunload                      //  The window, the document and its resources are about to be unloaded.
	HTMLEventunload                            //  The document or a dependent resource is being unloaded.

	// focus events
	HTMLEventfocus    // An element has received focus (does not bubble).
	HTMLEventblur     // An element has lost focus (does not bubble).
	HTMLEventfocusin  // An element is about to receive focus (does bubble).
	HTMLEventfocusout // An element is about to lose focus (does bubble).

	// css events
	HTMLEventanimationstart     // A CSS animation has started.
	HTMLEventanimationcancel    // A CSS animation has aborted.
	HTMLEventanimationend       // A CSS animation has completed.
	HTMLEventanimationiteration // A CSS animation is repeated.
	HTMLEventtransitionstart    // A CSS transition has actually started (fired after any delay).
	HTMLEventtransitioncancel   // A CSS transition has been cancelled.
	HTMLEventtransitionend      // A CSS transition has completed.
	HTMLEventtransitionrun      // A CSS transition has begun running (fired before any delay starts).

	//  form events
	HTMLEventreset  // The reset button is pressed
	HTMLEventsubmit // The submit button is pressed

	//  print events
	HTMLEventbeforeprint // The print dialog is opened
	HTMLEventafterprint  // The print dialog is closed

	//  text composition events
	HTMLEventcompositionstart  // The composition of a passage of text is prepared (similar to keydown for a keyboard input, but works with other inputs such as speech recognition).
	HTMLEventcompositionupdate // A character is added to a passage of text being composed.
	HTMLEventcompositionend    // The composition of a passage of text has been completed or canceled.

	// view events
	HTMLEventfullscreenchange // An element was toggled to or from fullscreen mode.
	HTMLEventfullscreenerror  // It was impossible to switch to fullscreen mode for technical reasons or because the permission was denied.
	HTMLEventresize           // The document view has been resized.
	HTMLEventscroll           // The document view or an element has been scrolled.

	//  clipboard events
	HTMLEventcut   // The selection has been cut and copied to the clipboard
	HTMLEventcopy  // The selection has been copied to the clipboard
	HTMLEventpaste // The item from the clipboard has been pasted

	// keyboard events
	HTMLEventkeydown  // ANY key is pressed
	HTMLEventkeypress // ANY key (except Shift, Fn, or CapsLock) is in pressed position. (Fired continously.)
	HTMLEventkeyup    // ANY key is released

	//  mouse events
	HTMLEventauxclick          // A pointing device button (ANY non-primary button) has been pressed and released on an element.
	HTMLEventclick             // A pointing device button (ANY button; soon to be primary button only) has been pressed and released on an element.
	HTMLEventcontextmenu       // The right button of the mouse is clicked (before the context menu is displayed).
	HTMLEventdblclick          // A pointing device button is clicked twice on an element.
	HTMLEventmousedown         // A pointing device button is pressed on an element.
	HTMLEventmouseenter        // A pointing device is moved onto the element that has the listener attached.
	HTMLEventmouseleave        // A pointing device is moved off the element that has the listener attached.
	HTMLEventmousemove         // A pointing device is moved over an element. (Fired continously as the mouse moves.)
	HTMLEventmouseover         // A pointing device is moved onto the element that has the listener attached or onto one of its children.
	HTMLEventmouseout          // A pointing device is moved off the element that has the listener attached or off one of its children.
	HTMLEventmouseup           // A pointing device button is released over an element.
	HTMLEventpointerlockchange // The pointer was locked or released.
	HTMLEventpointerlockerror  // It was impossible to lock the pointer for technical reasons or because the permission was denied.
	HTMLEventselect            // Some text is being selected.
	HTMLEventwheel             // A wheel button of a pointing device is rotated in any direction.

	//  drag & drop events
	HTMLEventdrag      // An element or text selection is being dragged. (Fired continuously every 350ms)
	HTMLEventdragend   // A drag operation is being ended (by releasing a mouse button or hitting the escape key).
	HTMLEventdragenter // A dragged element or text selection enters a valid drop target.
	HTMLEventdragstart // The user starts dragging an element or text selection.
	HTMLEventdragleave // A dragged element or text selection leaves a valid drop target.
	HTMLEventdragover  // An element or text selection is being dragged over a valid drop target. (Fired continuously every 350ms)
	HTMLEventdrop      // An element is dropped on a valid drop target.

	// media events
	HTMLEventaudioprocess   // The input buffer of a ScriptProcessorNode is ready to be processed.
	HTMLEventcanplay        // The browser can play the media, but estimates that not enough data has been loaded to play the media up to its end without having to stop for further buffering of content.
	HTMLEventcanplaythrough // The browser estimates it can play the media up to its end without stopping for content buffering.
	HTMLEventcomplete       // The rendering of an OfflineAudioContext is terminated.
	HTMLEventdurationchange // The duration attribute has been updated.
	HTMLEventemptied        // The media has become empty; for example, this event is sent if the media has already been loaded (or partially loaded), and the load() method is called to reload it.
	HTMLEventended          // Playback has stopped because the end of the media was reached.
	HTMLEventloadeddata     // The first frame of the media has finished loading.
	HTMLEventloadedmetadata // The metadata has been loaded.
	HTMLEventpause          // Playback has been paused.
	HTMLEventplay           // Playback has begun.
	HTMLEventplaying        // Playback is ready to start after having been paused or delayed due to lack of data.
	HTMLEventratechange     // The playback rate has changed.
	HTMLEventseeked         // A seek operation completed.
	HTMLEventseeking        // A seek operation began.
	HTMLEventstalled        // The user agent is trying to fetch media data, but data is unexpectedly not forthcoming.
	HTMLEventsuspend        // Media data loading has been suspended.
	HTMLEventtimeupdate     // The time indicated by the currentTime attribute has been updated.
	HTMLEventvolumechange   // The volume has changed.
	HTMLEventwaiting        // Playback has stopped because of a temporary lack of data.

	// progress events
	HTMLEventloadend   // Progress has stopped (after "error", "abort", or "load" have been dispatched).
	HTMLEventloadstart // Progress has begun.
	HTMLEventprogress  // In progress.
	HTMLEventtimeout   // Progression is terminated due to preset time expiring.

	// touch events
	HTMLEventtouchcancel
	HTMLEventtouchend
	HTMLEventtouchmove
	HTMLEventtouchstart

	// pointer events
	HTMLEventpointerover
	HTMLEventpointerenter
	HTMLEventpointerdown
	HTMLEventpointermove
	HTMLEventpointerup
	HTMLEventpointercancel
	HTMLEventpointerout
	HTMLEventpointerleave
	HTMLEventgotpointercapture
	HTMLEventlostpointercapture

	// Custom/non-standard
	HTMLEventcustom
)

// HTMLEventTypeMap event type -> type
var HTMLEventTypeMap = map[string]HTMLEventType{
	"error":        HTMLEventerror,
	"abort":        HTMLEventabort,
	"load":         HTMLEventload,
	"beforeunload": HTMLEventbeforeunload,
	"unload":       HTMLEventunload,

	// focus events
	"focus":    HTMLEventfocus,
	"blur":     HTMLEventblur,
	"focusin":  HTMLEventfocusin,
	"focusout": HTMLEventfocusout,

	// css events
	"animationstart":     HTMLEventanimationstart,
	"animationcancel":    HTMLEventanimationcancel,
	"animationend":       HTMLEventanimationend,
	"animationiteration": HTMLEventanimationiteration,
	"transitionstart":    HTMLEventtransitionstart,
	"transitioncancel":   HTMLEventtransitioncancel,
	"transitionend":      HTMLEventtransitionend,
	"transitionrun":      HTMLEventtransitionrun,

	//  form events
	"reset":  HTMLEventreset,
	"submit": HTMLEventsubmit,

	//  print events
	"beforeprint": HTMLEventbeforeprint,
	"afterprint":  HTMLEventafterprint,

	//  text composition events
	"compositionstart":  HTMLEventcompositionstart,
	"compositionupdate": HTMLEventcompositionupdate,
	"compositionend":    HTMLEventcompositionend,

	// view events
	"fullscreenchange": HTMLEventfullscreenchange,
	"fullscreenerror":  HTMLEventfullscreenerror,
	"resize":           HTMLEventresize,
	"scroll":           HTMLEventscroll,

	//  clipboard events
	"cut":   HTMLEventcut,
	"copy":  HTMLEventcopy,
	"paste": HTMLEventpaste,

	// keyboard events
	"keydown":  HTMLEventkeydown,
	"keypress": HTMLEventkeypress,
	"keyup":    HTMLEventkeyup,

	//  mouse events
	"auxclick":          HTMLEventauxclick,
	"click":             HTMLEventclick,
	"contextmenu":       HTMLEventcontextmenu,
	"dblclick":          HTMLEventdblclick,
	"mousedown":         HTMLEventmousedown,
	"mouseenter":        HTMLEventmouseenter,
	"mouseleave":        HTMLEventmouseleave,
	"mousemove":         HTMLEventmousemove,
	"mouseover":         HTMLEventmouseover,
	"mouseout":          HTMLEventmouseout,
	"mouseup":           HTMLEventmouseup,
	"pointerlockchange": HTMLEventpointerlockchange,
	"pointerlockerror":  HTMLEventpointerlockerror,
	"select":            HTMLEventselect,
	"wheel":             HTMLEventwheel,

	//  drag & drop events
	"drag":      HTMLEventdrag,
	"dragend":   HTMLEventdragend,
	"dragenter": HTMLEventdragenter,
	"dragstart": HTMLEventdragstart,
	"dragleave": HTMLEventdragleave,
	"dragover":  HTMLEventdragover,
	"drop":      HTMLEventdrop,

	// media events
	"audioprocess":   HTMLEventaudioprocess,
	"canplay":        HTMLEventcanplay,
	"canplaythrough": HTMLEventcanplaythrough,
	"complete":       HTMLEventcomplete,
	"durationchange": HTMLEventdurationchange,
	"emptied":        HTMLEventemptied,
	"ended":          HTMLEventended,
	"loadeddata":     HTMLEventloadeddata,
	"loadedmetadata": HTMLEventloadedmetadata,
	"pause":          HTMLEventpause,
	"play":           HTMLEventplay,
	"playing":        HTMLEventplaying,
	"ratechange":     HTMLEventratechange,
	"seeked":         HTMLEventseeked,
	"seeking":        HTMLEventseeking,
	"stalled":        HTMLEventstalled,
	"suspend":        HTMLEventsuspend,
	"timeupdate":     HTMLEventtimeupdate,
	"volumechange":   HTMLEventvolumechange,
	"waiting":        HTMLEventwaiting,

	// progress events
	"loadend":   HTMLEventloadend,
	"loadstart": HTMLEventloadstart,
	"progress":  HTMLEventprogress,
	"timeout":   HTMLEventtimeout,

	// touch events
	"touchcancel": HTMLEventtouchcancel,
	"touchend":    HTMLEventtouchend,
	"touchmove":   HTMLEventtouchmove,
	"touchstart":  HTMLEventtouchstart,

	// pointer events
	"pointerover":        HTMLEventpointerover,
	"pointerenter":       HTMLEventpointerenter,
	"pointerdown":        HTMLEventpointerdown,
	"pointermove":        HTMLEventpointermove,
	"pointerup":          HTMLEventpointerup,
	"pointercancel":      HTMLEventpointercancel,
	"pointerout":         HTMLEventpointerout,
	"pointerleave":       HTMLEventpointerleave,
	"gotpointercapture":  HTMLEventgotpointercapture,
	"lostpointercapture": HTMLEventlostpointercapture,
}
