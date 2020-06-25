package mock

import (
	"context"
	"net/url"

	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

func MakeMockContext(ctx context.Context, target *url.URL) *browserk.Context {
	logger := log.With().
		Str("DEBUGURL", target.String()).
		Logger()
	return &browserk.Context{
		Log:            &logger,
		Ctx:            ctx,
		PluginServicer: MakeMockPluginServicer(),
		Auth:           nil,
		Scope:          MakeMockScopeService(target),
		FormHandler:    nil,
		Reporter:       nil,
		Crawl:          nil,
	}
}
