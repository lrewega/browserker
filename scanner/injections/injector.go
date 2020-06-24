package injections

import (
	"bytes"
	"context"
	"fmt"
	"time"

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
		browser:     browser,
		msgIterator: msgIterator,
		injIterator: injIterator,
		bCtx:        bCtx,
	}
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

func (i *BrowserkerInjector) Send(ctx context.Context) (*browserk.HTTPResponse, error) {
	//i.injIterator.URI()
	if i.msgIterator.Request().Type == "Document" {
		// inject <form>

		i.bCtx.Log.Debug().Msg("injecting form")
	} else {
		// inject xhr
		i.bCtx.AddReqHandler()
		i.bCtx.Log.Debug().Msg("injecting xhr")
	}
	return nil, nil
}

func (i *BrowserkerInjector) SendNew(ctx context.Context, req *browserk.HTTPRequest) (*browserk.InterceptedHTTPResponse, error) {
	//i.injIterator.URI()
	if i.msgIterator.Request().Type == "Document" {
		// inject <form>

		i.bCtx.Log.Debug().Msg("injecting form")
	} else {
		// inject xhr
		respCh := make(chan *browserk.InterceptedHTTPResponse)
		i.bCtx.AddReqHandler(InjectXHRReq(respCh, req.Request.Url, req.Request.Headers, req.Request.PostData, i.req.Hash()))
		i.bCtx.Log.Debug().Msg("injecting xhr")
		//i.browser.
		timer := time.NewTimer(time.Minute * 2)
		select {
		case r := <-respCh:
			return r, nil
		case <-timer.C:
			return nil, fmt.Errorf("failed to get response from injection")
		}
	}
	return nil, nil
}

func InjectXHRReq(respCh chan *browserk.InterceptedHTTPResponse, newURI string, headers map[string]interface{}, body string, match []byte) browserk.RequestHandler {
	return func(bctx *browserk.Context, browser browserk.Browser, i *browserk.InterceptedHTTPRequest) {
		if bytes.Compare(i.Hash(), match) != 0 {
			return
		}
		i.Modified.Url = newURI
		i.Modified.SetHeaders(headers)
		i.Modified.PostData = body
		bctx.AddRespHandler(func(respBctx *browserk.Context, respBrowser browserk.Browser, resp *browserk.InterceptedHTTPResponse) {
			if resp.RequestId == i.RequestId {
				select {
				case <-respBctx.Ctx.Done():
					return
				case respCh <- resp:
				}
			}
		})
	}
}
