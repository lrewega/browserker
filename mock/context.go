package mock

import (
	"context"
	"net/url"

	"github.com/rs/zerolog"
	"gitlab.com/browserker/browserk"
)

func MakeMockContext(ctx context.Context, target *url.URL) *browserk.Context {
	log := &zerolog.Logger{}
	return &browserk.Context{
		Log:            log,
		Ctx:            ctx,
		PluginServicer: MakeMockPluginServicer(),
		Auth:           nil,
		Scope:          MakeMockScopeService(target),
		FormHandler:    nil,
		Reporter:       nil,
		Injector:       nil,
		Crawl:          nil,
	}
}
