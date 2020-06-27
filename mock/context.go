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
	bctx := browserk.NewContext(ctx, nil)
	bctx.Log = &logger
	bctx.PluginServicer = MakeMockPluginServicer()
	bctx.Scope = MakeMockScopeService(target)
	return bctx
}
