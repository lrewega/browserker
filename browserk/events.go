package browserk

import "time"

// StorageEventType of storage related events.
type StorageEventType int8

const (
	// StorageClearedEvt cleared the storage type
	StorageClearedEvt StorageEventType = iota
	// StorageRemovedEvt removed an element
	StorageRemovedEvt
	// StorageAddedEvt added an element
	StorageAddedEvt
	// StorageUpdatedEvt updated an element
	StorageUpdatedEvt
)

// StorageEvent captures local/sessionStorage details
type StorageEvent struct {
	Type           StorageEventType `json:"type"`      // Type of storage event
	IsLocalStorage bool             `json:"is_local"`  // if true, local storage, false session storage
	SecurityOrigin string           `json:"origin"`    // origin that this event occurred on
	Key            string           `json:"key"`       // storage key
	NewValue       string           `json:"new_value"` // new storage value
	OldValue       string           `json:"old_value"` // old storage value
	Observed       time.Time        `json:"observed"`  // time the storage event occurred
}

// ConsoleEvent captures console.log events
type ConsoleEvent struct {
	Source   string    `json:"source"`           // Message source.
	Level    string    `json:"level"`            // Message severity.
	Text     string    `json:"text"`             // Message text.
	URL      string    `json:"url,omitempty"`    // URL of the message origin.
	Line     int       `json:"line,omitempty"`   // Line number in the resource that generated this message (1-based).
	Column   int       `json:"column,omitempty"` // Column number in the resource that generated this message (1-based).
	Observed time.Time `json:"observed"`         // time the console event occurred
}
