package scanner

import (
	"context"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/auth"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
	"gitlab.com/browserker/scanner/report"
)

// Browserk is our engine
type Browserk struct {
	cfg          *browserk.Config
	attackGraph  browserk.AttackGrapher
	crawlGraph   browserk.CrawlGrapher
	reporter     browserk.Reporter
	browsers     browserk.BrowserPool
	navEntryCh   chan [][]*browserk.Navigation
	readyCh      chan struct{}
	stateMonitor *time.Ticker
	mainContext  *browserk.Context
}

// New engine
func New(cfg *browserk.Config, crawl browserk.CrawlGrapher, attack browserk.AttackGrapher) *Browserk {
	return &Browserk{cfg: cfg, attackGraph: attack, crawlGraph: crawl, reporter: report.New()}
}

// SetReporter overrides the default reporter
func (b *Browserk) SetReporter(reporter browserk.Reporter) *Browserk {
	b.reporter = reporter
	return b
}

// Init the browsers and stores
func (b *Browserk) Init(ctx context.Context) error {
	target, err := url.Parse(b.cfg.URL)
	if err != nil {
		return err
	}
	cancelCtx, cancelFn := context.WithCancel(ctx)

	b.mainContext = &browserk.Context{
		Ctx:         cancelCtx,
		CtxComplete: cancelFn,
		Auth:        auth.New(b.cfg),
		Scope:       b.scopeService(target),
		Reporter:    b.reporter,
		Injector:    nil,
		Crawl:       b.crawlGraph,
		Attack:      b.attackGraph,
	}

	b.navEntryCh = make(chan [][]*browserk.Navigation, b.cfg.NumBrowsers)
	b.readyCh = make(chan struct{})

	log.Logger.Info().Msg("initializing attack graph")
	if err := b.attackGraph.Init(); err != nil {
		return err
	}

	log.Logger.Info().Msg("initializing crawl graph")
	if err := b.crawlGraph.Init(); err != nil {
		return err
	}

	b.initNavigation()

	b.stateMonitor = time.NewTicker(time.Second * 10)

	log.Logger.Info().Msg("starting leaser")
	leaser := browser.NewLocalLeaser()
	log.Logger.Info().Msg("leaser started")
	pool := browser.NewGCDBrowserPool(b.cfg.NumBrowsers, leaser)
	b.browsers = pool
	log.Logger.Info().Msg("starting browser pool")
	go b.processEntries()
	return pool.Init()
}

func (b *Browserk) initNavigation() {
	nav := browserk.NewNavigation(browserk.TrigInitial, &browserk.Action{
		Type:   browserk.ActLoadURL,
		Input:  []byte(b.cfg.URL),
		Result: nil,
	})
	nav.Scope = browserk.InScope
	nav.Distance = 0

	// reset any inprocess navigations to unvisited because it didn't exit cleanly
	b.crawlGraph.Find(b.mainContext.Ctx, browserk.NavInProcess, browserk.NavUnvisited, 1000)

	if !b.crawlGraph.NavExists(nav) {
		b.crawlGraph.AddNavigation(nav)
		log.Info().Msg("Load URL added to crawl graph")
	} else {
		log.Info().Msg("Navigation for Load URL already exists")
	}
}

func (b *Browserk) scopeService(target *url.URL) browserk.ScopeService {
	allowed := b.cfg.AllowedHosts
	ignored := b.cfg.IgnoredHosts
	excluded := b.cfg.ExcludedHosts

	if allowed == nil {
		allowed = []string{target.Hostname()}
	}

	scope := NewScopeService(target)
	scope.AddScope(allowed, browserk.InScope)
	scope.AddScope(ignored, browserk.OutOfScope)
	scope.AddScope(excluded, browserk.ExcludedFromScope)
	if b.cfg.ExcludedURIs != nil {
		scope.AddExcludedURIs(b.cfg.ExcludedURIs)
	}
	return scope
}

// Start the browsers
func (b *Browserk) Start() error {
	for {
		log.Info().Msg("searching for new navigation entries")
		entries := b.crawlGraph.Find(b.mainContext.Ctx, browserk.NavUnvisited, browserk.NavInProcess, int64(b.cfg.NumBrowsers))
		if entries == nil || len(entries) == 0 && b.browsers.Leased() == 0 {
			log.Info().Msg("no more crawler entries or active browsers")
			return nil
		}
		log.Info().Int("entries", len(entries)).Msg("Found entries")
		b.navEntryCh <- entries
		log.Info().Msg("Waiting for crawler to complete")
		<-b.readyCh
	}
}

func (b *Browserk) processEntries() {
	for {
		select {
		case <-b.stateMonitor.C:
			// TODO: check graph for inprocess values that never made it
			log.Info().Msg("state monitor ping")
		case <-b.mainContext.Ctx.Done():
			log.Info().Msg("scan finished due to context complete")
			return
		case entries := <-b.navEntryCh:
			if err := b.crawl(entries); err != nil {
				log.Error().Err(err).Msg("crawl entries failed")
			}
			log.Info().Msg("Crawler to complete")
			b.readyCh <- struct{}{}
		}
	}
}

func (b *Browserk) crawl(entries [][]*browserk.Navigation) error {
	navCtx := b.mainContext.Copy()

	for _, navs := range entries {
		browser, port, err := b.browsers.Take(navCtx)
		if err != nil {
			return err
		}

		crawler := crawler.New(b.cfg)
		if err := crawler.Init(); err != nil {
			b.browsers.Return(navCtx.Ctx, port)
			return errors.Wrap(err, "failed to init crawler")
		}

		for i, nav := range navs {
			isFinal := false
			ctx, cancel := context.WithTimeout(navCtx.Ctx, time.Second*45)
			navCtx.Ctx = ctx
			defer cancel()

			// we are on the last navigation of this path so we'll want to capture some stuff
			if i == len(navs)-1 {
				isFinal = true
			}

			result, newNavs, err := crawler.Process(navCtx, browser, nav, isFinal)
			if err != nil {
				log.Error().Err(err).Msg("failed to process action")
			}

			if err := b.crawlGraph.AddNavigations(newNavs); err != nil {
				log.Error().Err(err).Msg("failed to add new navigations")
			}

			if err := b.crawlGraph.AddResult(result); err != nil {
				log.Error().Err(err).Msg("failed to add result")
			}
		}
		b.browsers.Return(navCtx.Ctx, port)
	}
	return nil
}

// Stop the browsers
func (b *Browserk) Stop() error {

	log.Info().Msg("Completing Ctx")
	b.mainContext.CtxComplete()

	log.Info().Msg("Stopping browsers")
	err := b.browsers.Shutdown()
	if err != nil {
		log.Warn().Err(err).Msg("failed to close browsers")
	}

	log.Info().Msg("Closing attack graph")
	err = b.attackGraph.Close()
	if err != nil {
		log.Warn().Err(err).Msg("failed to close attackGraph")
	}

	log.Info().Msg("Closing crawl graph")
	err = b.crawlGraph.Close()
	if err != nil {
		log.Warn().Err(err).Msg("failed to close crawlGraph")
	}
	return err
}
