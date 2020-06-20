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

func (b *BrowserkAttacker) Attack(bctx *browserk.Context, browser browserk.Browser, entry *browserk.NavigationWithResult, isFinal bool) (*browserk.NavigationResult, error) {
	// execute the action
	navCtx, cancel := context.WithTimeout(bctx.Ctx, time.Second*45)
	defer cancel()

	_, _, err := browser.ExecuteAction(navCtx, entry.Navigation)
	if err != nil {
		return nil, err
	}
	// TODO: iterate over requests (need to filter duplicates)
	//bctx.PluginServicer.
	return nil, nil
}
