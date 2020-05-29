package browserk

import (
	"crypto/md5"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// TriggeredBy stores what caused a navigation attempt
type TriggeredBy int16

const (
	// TrigInitial triggered action (for example start of crawl load url)
	TrigInitial = iota + 1
	// TrigCrawler triggered this navigation
	TrigCrawler
	// TrigPlugin triggered this
	TrigPlugin
	// TrigAutoBrowser something caused the browser to trigger this (redirect etc)
	TrigAutoBrowser
)

// NavState is the state of a navigation
type NavState int8

const (
	// NavInvalid is invalid
	NavInvalid NavState = iota + 1
	// NavUnvisited means it's ready for pick up by crawler
	NavUnvisited
	// NavInProcess crawler is in the process of crawling this
	NavInProcess
	// NavVisited crawler has visited
	NavVisited
	// NavFailed unable to complete action
	NavFailed
)

// Navigation for storing the action and results of navigating
type Navigation struct {
	ID               []byte      `graph:"id"`            // unique id of this navigation depending on type
	OriginID         []byte      `graph:"origin"`        // where this navigation node originated from
	TriggeredBy      TriggeredBy `graph:"trig_by"`       // update to plugin/crawler/manual whatever type
	State            NavState    `graph:"state"`         // state of this navigation
	StateUpdatedTime time.Time   `graph:"state_updated"` // when the state was updated (for timeouts)
	Action           *Action     `graph:"action"`
	Scope            Scope       `graph:"scope"`
	Distance         int         `graph:"dist"`
}

// NewNavigation type
func NewNavigation(triggeredBy TriggeredBy, action *Action) *Navigation {
	n := &Navigation{
		Action:           action,
		TriggeredBy:      triggeredBy,
		State:            NavUnvisited,
		StateUpdatedTime: time.Now(),
		Scope:            InScope,
	}

	// TODO: add originID as part of new nav id for uniqueness?
	h := md5.New()
	h.Write(n.Action.Input)
	h.Write([]byte{byte(n.Action.Type)})
	n.ID = h.Sum(nil)
	log.Info().Msgf("NEW ID: %#v", n.ID)
	return n
}

// NewNavigationFromForm creates a new navigation entry from forms
func NewNavigationFromForm(from *Navigation, triggeredBy TriggeredBy, form *HTMLFormElement) *Navigation {

	action := &Action{
		Type:    ActFillForm,
		Input:   nil,
		Element: nil,
		Form:    form,
		Result:  nil,
	}

	n := &Navigation{
		Action:           action,
		OriginID:         from.ID,
		TriggeredBy:      triggeredBy,
		State:            NavUnvisited,
		StateUpdatedTime: time.Now(),
		Scope:            InScope,
		Distance:         from.Distance + 1,
	}

	h := md5.New()
	h.Write(n.OriginID)
	h.Write(n.Action.Form.Hash())
	n.ID = h.Sum(nil)
	return n
}

// NewNavigationFromElement creates a new navigation entry from eventable elements
func NewNavigationFromElement(from *Navigation, triggeredBy TriggeredBy, ele *HTMLElement, aType ActionType) *Navigation {

	action := &Action{
		Type:    aType,
		Input:   nil,
		Element: ele,
		Form:    nil,
		Result:  nil,
	}

	n := &Navigation{
		Action:           action,
		OriginID:         from.ID,
		TriggeredBy:      triggeredBy,
		State:            NavUnvisited,
		StateUpdatedTime: time.Now(),
		Scope:            InScope,
		Distance:         from.Distance + 1,
	}

	h := md5.New()
	// we only want uniqueness of origin id's for links that would be unique on a page
	// we don't want to keep going to /page if it exists on *every* page.
	if link, ok := ele.Attributes["href"]; ok && ele.Type == A {
		if strings.HasPrefix(link, "#") {
			h.Write(n.OriginID)
		}
	}
	h.Write([]byte{byte(aType)})
	h.Write(n.Action.Element.Hash())
	n.ID = h.Sum(nil)
	return n
}

// NavigationResult captures result details about a navigation
type NavigationResult struct {
	ID            []byte          `graph:"r_id"`
	NavigationID  []byte          `graph:"r_nav_id"`
	DOM           string          `graph:"r_dom"`
	StartURL      string          `graph:"r_start_url"`
	EndURL        string          `graph:"r_end_url"`
	MessageCount  int             `graph:"r_message_count"`
	Messages      []*HTTPMessage  `graph:"r_messages"`
	Cookies       []*Cookie       `graph:"r_cookies"`
	ConsoleEvents []*ConsoleEvent `graph:"r_console"`
	StorageEvents []*StorageEvent `graph:"r_storage"`
	CausedLoad    bool            `graph:"r_caused_load"`
	WasError      bool            `graph:"r_was_error"`
	Errors        []error         `graph:"r_errors"`
}

// Hash a unique ID for this result (needs work)
func (n *NavigationResult) Hash() []byte {
	if n.ID != nil {
		return n.ID
	}
	h := md5.New()
	// TODO come up with something better
	h.Write(n.NavigationID)
	h.Write([]byte(n.StartURL))
	h.Write([]byte(n.EndURL))
	if n.MessageCount > 0 && n.Messages[0].Request != nil {
		h.Write([]byte(n.Messages[0].Request.DocumentURL))
	}
	n.ID = h.Sum(nil)
	return n.ID
}
func (n *NavigationResult) AddError(err error) {
	if err != nil {
		n.Errors = append(n.Errors, err)
	}
}
