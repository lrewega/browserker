package attack

import (
	"context"
	"time"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/store"
)

type BrowserkAttacker struct {
	cfg        *browserk.Config
	crawlGraph store.CrawlGraph
}

func New(cfg *browserk.Config) *BrowserkAttacker {
	return &BrowserkAttacker{cfg: cfg}
}

func (b *BrowserkAttacker) Init() error {
	return nil
}

func (b *BrowserkAttacker) Attack(bctx *browserk.Context, browser browserk.Browser, entry *browserk.Navigation, isFinal bool) (*browserk.NavigationResult, error) {
	// execute the action
	navCtx, cancel := context.WithTimeout(bctx.Ctx, time.Second*45)
	defer cancel()

	errors := make([]error, 0)
	startURL, err := browser.GetURL()
	if err != nil {
		errors = append(errors, err)
	}

	result := &browserk.NavigationResult{
		ID:           nil,
		NavigationID: entry.ID,
		StartURL:     startURL,
		Errors:       errors,
	}

	_, result.CausedLoad, err = browser.ExecuteAction(navCtx, entry)
	if err != nil {
		result.WasError = true
		return result, err
	}
	return nil, nil
}
