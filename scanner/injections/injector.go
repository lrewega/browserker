package injections

import (
	"context"

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
		i.bCtx.Log.Debug().Msg("injecting xhr")
	}
	return nil, nil
}

func (i *BrowserkerInjector) SendNew(ctx context.Context, req *browserk.HTTPRequest) (*browserk.HTTPResponse, error) {
	//i.injIterator.URI()
	if i.msgIterator.Request().Type == "Document" {
		// inject <form>

		i.bCtx.Log.Debug().Msg("injecting form")
	} else {
		// inject xhr
		i.bCtx.Log.Debug().Msg("injecting xhr")
	}
	return nil, nil
}
