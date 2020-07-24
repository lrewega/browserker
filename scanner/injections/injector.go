package injections

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/iterator"
)

type BrowserkerInjector struct {
	nav         *browserk.NavigationWithResult
	browser     browserk.Browser
	msgIterator *iterator.MessageIterator
	injIterator *iterator.InjectionIterator
	req         *browserk.HTTPRequest
	bCtx        *browserk.Context
}

func New(bCtx *browserk.Context, browser browserk.Browser, nav *browserk.NavigationWithResult, msgIterator *iterator.MessageIterator, injIterator *iterator.InjectionIterator) *BrowserkerInjector {
	return &BrowserkerInjector{
		nav:         nav,
		req:         msgIterator.Request(),
		browser:     browser,
		msgIterator: msgIterator,
		injIterator: injIterator,
		bCtx:        bCtx,
	}
}

func (i *BrowserkerInjector) Nav() *browserk.Navigation {
	return i.nav.Navigation
}

func (i *BrowserkerInjector) NavResultID() []byte {
	return i.nav.Result.Hash()
}

func (i *BrowserkerInjector) BCtx() *browserk.Context {
	return i.bCtx
}

func (i *BrowserkerInjector) Message() *browserk.HTTPMessage {
	return i.msgIterator.Message().Copy()
}

func (i *BrowserkerInjector) InjectionExpr() browserk.InjectionExpr {
	return i.injIterator.Expr()
}

func (i *BrowserkerInjector) Browser() browserk.Browser {
	return i.browser
}

func (i *BrowserkerInjector) ReplacePath(newValue string, index int) {}
func (i *BrowserkerInjector) ReplaceFile(newValue string)            {}
func (i *BrowserkerInjector) ReplaceURI(newURI string)               {}
func (i *BrowserkerInjector) ReplaceHeader(name, value string)       {}
func (i *BrowserkerInjector) AddHeader(name, value string)           {}
func (i *BrowserkerInjector) RemoveHeader(name string)               {}
func (i *BrowserkerInjector) ReplaceBody(newBody []byte)             {}

// Send this injection attack
func (i *BrowserkerInjector) Send(ctx context.Context, withRender bool) (*browserk.InterceptedHTTPMessage, error) {
	//i.injIterator.URI()
	if withRender {
		// inject <form>

		i.bCtx.Log.Debug().Msg("injecting form")
	} else {
		respCh := make(chan *browserk.InterceptedHTTPMessage)
		id := rand.Int63()
		attackID := fmt.Sprintf("/injection%d", id)

		host, _ := iterator.SplitHost(i.req.Request.Url)
		// TODO: replace headers with injIterator.Headers body with injIterator.Body (those three should be separate)
		i.bCtx.Log.Debug().Str("location", i.injIterator.Expr().Loc().String()).
			Str("attack_METHOD", i.injIterator.Method()).
			Str("attack_URL", host+i.injIterator.URI().String()).
			Str("attack_BODY", i.injIterator.SerializeBody()).
			Int64("attack_id", id).
			Msg("injecting attack")

		injectFn := InjectFetchReq(respCh, i.injIterator.Method(), host+i.injIterator.SerializeURI(), i.req.Request.Headers, i.injIterator.SerializeBody(), attackID)
		i.bCtx.AddReqHandler(injectFn)

		i.bCtx.Log.Debug().Int64("attack_id", id).Msg("injecting js fetch")
		i.injIterator.Expr().Reset() // un-inject ourselves

		// issue request to hijack
		if err := i.browser.InjectRequest(ctx, i.req.Request.Method, host+attackID); err != nil {
			i.bCtx.Log.Error().Err(err).Int64("attack_id", id).Msg("failed to inject fetch attack")
			return nil, fmt.Errorf("injection failed")
		}

		select {
		case r := <-respCh:
			i.bCtx.Log.Debug().Msg("got response from attack")
			return r, nil
		case <-ctx.Done():
			i.bCtx.Log.Error().Int64("attack_id", id).Msg("injection timeout")
			return nil, browserk.ErrInjectionTimeout
		}
	}
	return nil, nil
}

// SendNew request instead of the modified one
func (i *BrowserkerInjector) SendNew(ctx context.Context, req *browserk.HTTPRequest, withRender bool) (*browserk.InterceptedHTTPMessage, error) {
	//i.injIterator.URI()
	if withRender {
		// inject <form>

		i.bCtx.Log.Debug().Msg("injecting form")
	} else {
		// inject xhr
		id := rand.Int63()
		attackID := fmt.Sprintf("/injection%d", id)
		host, _ := iterator.SplitHost(req.Request.Url)

		respCh := make(chan *browserk.InterceptedHTTPMessage)
		i.bCtx.AddReqHandler(InjectFetchReq(respCh, req.Request.Method, req.Request.Url, req.Request.Headers, req.Request.PostData, attackID))
		i.bCtx.Log.Debug().Msg("injecting fetch attack")

		if err := i.browser.InjectRequest(ctx, "GET", host+attackID); err != nil {
			i.bCtx.Log.Error().Err(err).Msg("failed to inject fetch attack")
			return nil, errors.Wrap(err, "injection failed")
		}
		select {
		case r := <-respCh:

			return r, nil
		case <-ctx.Done():
			return nil, fmt.Errorf("failed to get response, context done")
		}
	}
	return nil, nil
}

// InjectFetchReq into the browser
func InjectFetchReq(respCh chan *browserk.InterceptedHTTPMessage, newMethod, newURI string, headers map[string]interface{}, body string, match string) browserk.RequestHandler {
	return func(bctx *browserk.Context, browser browserk.Browser, i *browserk.InterceptedHTTPRequest) bool {
		_, uri := iterator.SplitHost(i.Request.Url)
		bctx.Log.Debug().Str("intercept_uri", uri).Str("inject_url_id", match).Msg("intercepted")
		if uri != match {
			bctx.Log.Debug().Str("intercept_uri", uri).Str("inject_url_id", match).Msg("did not match attack request")
			return false
		}
		bctx.Log.Debug().Str("newURI", newURI).Msg("matched attack request, rewriting")
		i.Modified.Method = newMethod
		i.Modified.Url = newURI
		i.Modified.SetHeaders(headers)
		i.Modified.PostData = body
		i.SentTimestamp = time.Now()
		bctx.Log.Debug().Str("response_key", i.FrameId+i.NetworkId).Msg("registered for response")
		bctx.PluginServicer.RegisterForResponse(i.FrameId+i.NetworkId, respCh, i)
		return true
	}
}
