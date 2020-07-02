package crawler

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

// BrowserkCrawler crawls a site
type BrowserkCrawler struct {
	cfg *browserk.Config
}

// New crawler for a site
func New(cfg *browserk.Config) *BrowserkCrawler {
	return &BrowserkCrawler{cfg: cfg}
}

// Init the crawler, if necessary
func (b *BrowserkCrawler) Init() error {
	return nil
}

// Process the next navigation entry
func (b *BrowserkCrawler) Process(bctx *browserk.Context, browser browserk.Browser, entry *browserk.Navigation, isFinal bool) (*browserk.NavigationResult, []*browserk.Navigation, error) {
	diff := NewElementDiffer()

	errors := make([]error, 0)
	startURL, err := browser.GetURL()
	if err != nil {
		errors = append(errors, err)
	}
	startCookies, err := browser.GetCookies()

	//clear out storage and console events before executing our action
	browser.GetStorageEvents()
	browser.GetConsoleEvents()

	if isFinal {
		diff = b.snapshot(bctx, browser)
	}

	result := &browserk.NavigationResult{
		ID:           nil,
		NavigationID: entry.ID,
		StartURL:     startURL,
		Errors:       errors,
		Cookies:      startCookies,
	}

	// execute the action
	navCtx, cancel := context.WithTimeout(bctx.Ctx, time.Second*45)
	defer cancel()
	beforeAction := time.Now()
	_, result.CausedLoad, err = browser.ExecuteAction(navCtx, entry)
	if err != nil {
		result.WasError = true
		return result, nil, err
	}

	// capture results
	b.buildResult(result, beforeAction, browser)

	// dispatch new cookie event
	for _, cookie := range result.Cookies {
		bctx.PluginServicer.DispatchEvent(browserk.CookiePluginEvent(bctx, result.EndURL, entry, cookie))
	}

	// find new potential navigation entries (if isFinal)
	potentialNavs := make([]*browserk.Navigation, 0)
	if isFinal {
		potentialNavs = b.FindNewNav(bctx, diff, entry, browser)
	}
	return result, potentialNavs, nil
}

// buildResult captures various data points after we executed an Action
func (b *BrowserkCrawler) buildResult(result *browserk.NavigationResult, start time.Time, browser browserk.Browser) {
	messages, err := browser.GetMessages()
	result.AddError(err)
	result.Messages = browserk.MessagesAfterRequestTime(messages, start)
	result.MessageCount = len(result.Messages)
	dom, err := browser.GetDOM()
	result.AddError(err)
	result.DOM = dom
	endURL, err := browser.GetURL()
	result.AddError(err)
	result.EndURL = endURL
	cookies, err := browser.GetCookies()
	result.AddError(err)
	result.Cookies = browserk.DiffCookies(result.Cookies, cookies)
	result.StorageEvents = browser.GetStorageEvents()
	result.ConsoleEvents = browser.GetConsoleEvents()
	result.Hash()
}

func (b *BrowserkCrawler) snapshot(bctx *browserk.Context, browser browserk.Browser) *ElementDiffer {
	diff := NewElementDiffer()
	browser.RefreshDocument()
	baseHref := browser.GetBaseHref()

	if formElements, err := browser.FindForms(); err == nil {
		for _, ele := range formElements {
			diff.Add(browserk.FORM, ele.Hash())
			//for _, child := range ele.ChildElements {
			//diff.Add(child.Type, child.Hash())
			//}
		}
	}

	if bElements, err := browser.FindElements("button"); err == nil {
		for _, ele := range bElements {
			// we want events that make elements visible to be executed first, so don't add 'em yet
			if !ele.Hidden {
				diff.Add(browserk.BUTTON, ele.Hash())
			}
		}
	}

	if aElements, err := browser.FindElements("a"); err == nil {
		for _, ele := range aElements {
			// we want events that make elements visible to be executed first, so don't add 'em yet
			if ele.Hidden {
				continue
			}

			scope := bctx.Scope.ResolveBaseHref(baseHref, ele.Attributes["href"])
			if scope == browserk.InScope {
				diff.Add(browserk.A, ele.Hash())
			}
		}
	}

	if txtElements, err := browser.FindElements("#text"); err == nil {
		for _, ele := range txtElements {
			// we want events that make elements visible to be executed first, so don't add 'em yet
			if ele.Hidden {
				continue
			}
			diff.Add(browserk.HASHTEXT, ele.Hash())
		}
	}

	if imgElements, err := browser.FindElements("img"); err == nil {
		for _, ele := range imgElements {
			// we want events that make elements visible to be executed first, so don't add 'em yet
			if ele.Hidden {
				continue
			}
			diff.Add(browserk.IMG, ele.Hash())
		}
	}

	cElements, err := browser.FindInteractables()
	if err == nil {
		for _, ele := range cElements {
			// we want events that make elements visible to be executed first, so don't add 'em yet
			if ele.Hidden {
				continue
			}
			// assume in scope for now
			diff.Add(ele.Type, ele.Hash())
			if ele.Type == browserk.HTML {
				spew.Dump(ele)
			}
		}
	}
	return diff
}

// FindNewNav potentials TODO: get navigation entry metadata (is vuejs/react etc) to be more specific
func (b *BrowserkCrawler) FindNewNav(bctx *browserk.Context, diff *ElementDiffer, entry *browserk.Navigation, browser browserk.Browser) []*browserk.Navigation {
	navs := make([]*browserk.Navigation, 0)
	browser.RefreshDocument()
	baseHref := browser.GetBaseHref()

	// Pull out forms (highest priority)
	formElements, err := browser.FindForms()
	if err != nil {
		bctx.Log.Info().Err(err).Msg("error while extracting forms")
	}

	for _, form := range formElements {
		scope := bctx.Scope.ResolveBaseHref(baseHref, form.GetAttribute("action"))
		if scope == browserk.InScope && !diff.Has(browserk.FORM, form.Hash()) && !form.Hidden {
			nav := browserk.NewNavigationFromForm(entry, browserk.TrigCrawler, form)
			bctx.FormHandler.Fill(form)
			navs = append(navs, nav)
		} else {
			if form.Hidden {
				bctx.Log.Debug().Str("href", baseHref).Str("action", form.GetAttribute("action")).Msg("was hidden, not creating new nav")
			} else {
				bctx.Log.Debug().Str("href", baseHref).Str("action", form.GetAttribute("action")).Msg("was out of scope or already found, not creating new nav")
			}

		}
	}

	bElements, err := browser.FindElements("button")
	if err != nil {
		bctx.Log.Info().Err(err).Msg("error while extracting buttons")
	}

	for _, b := range bElements {
		if !diff.Has(browserk.BUTTON, b.Hash()) && !b.Hidden {
			bctx.Log.Info().Msgf("adding button %#v", b)
			navs = append(navs, browserk.NewNavigationFromElement(entry, browserk.TrigCrawler, b, browserk.ActLeftClick))
		}
	}

	aElements, err := browser.FindElements("a")
	if err != nil {
		bctx.Log.Error().Err(err).Msg("error while extracting links")
	} else if aElements == nil || len(aElements) == 0 {
		log.Warn().Msg("error while extracting links")
	}

	bctx.Log.Debug().Int("link_count", len(aElements)).Msg("found links")
	for _, a := range aElements {
		//href := a.GetAttribute("href")
		// skip a elements that don't have an href field, if they have event listeners bound, FindInteractables will pick them up.
		// for example <a name="something"> isn't worth clicking/going to.
		//if href == "" {
		//	continue
		//}
		scope := bctx.Scope.ResolveBaseHref(baseHref, a.GetAttribute("href"))
		if scope == browserk.InScope && !diff.Has(browserk.A, a.Hash()) && !a.Hidden {
			bctx.Log.Info().Str("baseHref", baseHref).Str("href", a.Attributes["href"]).Msg("in scope, adding")
			nav := browserk.NewNavigationFromElement(entry, browserk.TrigCrawler, a, browserk.ActLeftClick)
			nav.Scope = scope
			navs = append(navs, nav)
		} else {
			if a.Hidden {
				bctx.Log.Debug().Str("baseHref", baseHref).Str("linkHref", a.GetAttribute("href")).Msg("a element was hidden, not creating new nav")
			} else {
				bctx.Log.Debug().Str("baseHref", baseHref).Str("linkHref", a.GetAttribute("href")).Msg("a element was out of scope, not creating new nav")
			}
		}
	}

	textElements, err := browser.FindElements("#text")
	if err != nil {
		bctx.Log.Error().Err(err).Msg("error while extracting text")
	} else if aElements == nil || len(aElements) == 0 {
		log.Warn().Msg("error while extracting text")
	}

	for _, txt := range textElements {
		nav := browserk.NewNavigationFromElement(entry, browserk.TrigCrawler, txt, browserk.ActLeftClick)
		nav.Scope = browserk.InScope
		navs = append(navs, nav)
	}

	imgElements, err := browser.FindElements("img")
	if err != nil {
		bctx.Log.Error().Err(err).Msg("error while extracting images")
	} else if aElements == nil || len(aElements) == 0 {
		log.Warn().Msg("error while extracting images")
	}

	for _, img := range imgElements {
		nav := browserk.NewNavigationFromElement(entry, browserk.TrigCrawler, img, browserk.ActLeftClick)
		nav.Scope = browserk.InScope
		navs = append(navs, nav)
	}

	cElements, err := browser.FindInteractables()
	if err == nil {
		for _, ele := range cElements {
			// assume in scope for now
			if !diff.Has(ele.Type, ele.Hash()) && !ele.Hidden {
				for _, eventType := range ele.Events {
					var actType browserk.ActionType
					switch eventType {
					case browserk.HTMLEventfocusin, browserk.HTMLEventfocus:
						actType = browserk.ActFocus
					case browserk.HTMLEventblur, browserk.HTMLEventfocusout:
						actType = browserk.ActBlur
					case browserk.HTMLEventclick, browserk.HTMLEventauxclick, browserk.HTMLEventmousedown, browserk.HTMLEventmouseup:
						actType = browserk.ActLeftClick
					case browserk.HTMLEventdblclick:
						actType = browserk.ActDoubleClick
					case browserk.HTMLEventmouseover, browserk.HTMLEventmouseenter, browserk.HTMLEventmouseleave, browserk.HTMLEventmouseout:
						actType = browserk.ActMouseOverAndOut
					case browserk.HTMLEventkeydown, browserk.HTMLEventkeypress, browserk.HTMLEventkeyup:
						actType = browserk.ActSendKeys
					case browserk.HTMLEventwheel:
						actType = browserk.ActMouseWheel
					case browserk.HTMLEventcontextmenu:
						actType = browserk.ActRightClick
					}

					if actType == 0 {
						continue
					}
					log.Info().Msgf("Adding action: %s for eventType: %v", browserk.ActionTypeMap[actType], eventType)
					nav := browserk.NewNavigationFromElement(entry, browserk.TrigCrawler, ele, actType)
					nav.Scope = browserk.InScope
					log.Info().Msgf("nav hash: %s", string(nav.ID))
					navs = append(navs, nav)
				}
			} else {
				if ele.Hidden {
					bctx.Log.Debug().Str("ele", browserk.HTMLTypeToStrMap[ele.Type]).Msgf("this element was hidden %+v", ele.Attributes)
				} else {
					bctx.Log.Debug().Str("ele", browserk.HTMLTypeToStrMap[ele.Type]).Msgf("this element already exists %+v", ele.Attributes)
				}
			}
		}
	}
	return navs
}
