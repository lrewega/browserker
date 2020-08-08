package browserk

import (
	"context"
)

// BrowserPool handles taking/returning browsers
type BrowserPool interface {
	Take(ctx *Context) (Browser, string, error)
	Return(ctx context.Context, browserPort string)
	Leased() int
	Shutdown() error
}

// BrowserOpts todo: define
type BrowserOpts struct {
}

// Browser interface
type Browser interface {
	ID() int64
	Init(*Config) error
	GetURL() (string, error)
	GetDOM() (string, error)
	GetCookies() ([]*Cookie, error)
	GetBaseHref() string
	GetStorageEvents() []*StorageEvent
	GetConsoleEvents() []*ConsoleEvent
	Navigate(ctx context.Context, url string) (err error)
	FindElements(ctx context.Context, querySelector string, canRefreshDoc bool) ([]*HTMLElement, error)
	FindForms(ctx context.Context) ([]*HTMLFormElement, error)
	FindInteractables() ([]*HTMLElement, error)
	GetMessages() ([]*HTTPMessage, error)
	Screenshot() (string, error)
	InjectRequest(ctx context.Context, method, URI string) error
	RefreshDocument()                                                         // reloads the document/elements
	ExecuteAction(ctx context.Context, nav *Navigation) ([]byte, bool, error) // result, caused page load, err
	Close()
}
