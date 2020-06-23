package scanner

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/auth"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
	"gitlab.com/browserker/scanner/iterator"
	"gitlab.com/browserker/scanner/plugin"
	"gitlab.com/browserker/scanner/report"
)

// Browserk is our engine
type Browserk struct {
	cfg          *browserk.Config
	pluginStore  browserk.PluginStorer
	crawlGraph   browserk.CrawlGrapher
	reporter     browserk.Reporter
	browsers     browserk.BrowserPool
	formHandler  browserk.FormHandler
	navCh        chan []*browserk.Navigation
	attackCh     chan []*browserk.NavigationWithResult
	readyCh      chan struct{}
	stateMonitor *time.Ticker
	mainContext  *browserk.Context

	idMutex          *sync.RWMutex
	leasedBrowserIDs map[int64]struct{}
}

// New engine
func New(cfg *browserk.Config, crawl browserk.CrawlGrapher, pluginStore browserk.PluginStorer) *Browserk {
	return &Browserk{
		cfg:              cfg,
		pluginStore:      pluginStore,
		crawlGraph:       crawl,
		leasedBrowserIDs: make(map[int64]struct{}),
		idMutex:          &sync.RWMutex{},
		navCh:            make(chan []*browserk.Navigation, cfg.NumBrowsers),
		attackCh:         make(chan []*browserk.NavigationWithResult, cfg.NumBrowsers),
		reporter:         report.New(crawl, pluginStore),
		readyCh:          make(chan struct{}),
	}
}

// SetReporter overrides the default reporter
func (b *Browserk) SetReporter(reporter browserk.Reporter) *Browserk {
	b.reporter = reporter
	return b
}

func (b *Browserk) addLeased(id int64) {
	b.idMutex.Lock()
	b.leasedBrowserIDs[id] = struct{}{}
	b.idMutex.Unlock()
}

func (b *Browserk) removeLeased(id int64) {
	b.idMutex.Lock()
	delete(b.leasedBrowserIDs, id)
	b.idMutex.Unlock()
}

func (b *Browserk) getLeased() []int64 {
	leased := make([]int64, 0)
	b.idMutex.RLock()
	for k := range b.leasedBrowserIDs {
		leased = append(leased, k)
	}
	b.idMutex.RUnlock()
	return leased
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
	}

	pluginService := plugin.New(b.cfg, b.pluginStore)
	if err := pluginService.Init(ctx); err != nil {
		return err
	}

	b.mainContext.Auth = auth.New(b.cfg)
	b.mainContext.Scope = b.scopeService(target)
	b.mainContext.FormHandler = crawler.NewCrawlerFormHandler(b.cfg.FormData)
	b.mainContext.Reporter = b.reporter
	b.mainContext.Crawl = b.crawlGraph
	b.mainContext.PluginServicer = pluginService

	log.Info().Int("num_browsers", b.cfg.NumBrowsers).Int("max_depth", b.cfg.MaxDepth).Msg("Initializing...")

	log.Logger.Info().Msg("initializing attack graph")
	if err := b.pluginStore.Init(); err != nil {
		return err
	}

	log.Logger.Info().Msg("initializing crawl graph")
	if err := b.crawlGraph.Init(); err != nil {
		return err
	}

	b.reporter = report.New(b.crawlGraph, b.pluginStore)
	b.formHandler = crawler.NewCrawlerFormHandler(b.cfg.FormData)

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
	log.Info().Msgf("ADDING URL %s", b.cfg.URL)
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
			log.Info().Msg("no more crawler entries or active browsers, activating attack phase")
			break

		}
		log.Info().Int("entries", len(entries)).Msg("Found entries")
		for _, nav := range entries {
			b.navCh <- nav
		}
		log.Info().Msg("Waiting for crawler to complete")
		<-b.readyCh
	}

	for {
		entries := b.crawlGraph.FindWithResults(b.mainContext.Ctx, browserk.NavVisited, browserk.NavInProcess, int64(b.cfg.NumBrowsers))
		if entries == nil || len(entries) == 0 && b.browsers.Leased() == 0 {
			log.Info().Msg("no more crawler entries or active browsers, scan complete")
			return nil
		}
		log.Info().Int("entries", len(entries)).Msg("Found entries")
		for _, nav := range entries {
			b.attackCh <- nav
		}
		log.Info().Msg("Waiting for crawler to complete")
		<-b.readyCh
	}
}

func (b *Browserk) processEntries() {
	for {
		select {
		case <-b.stateMonitor.C:
			// TODO: check graph for inprocess values that never made it and reset them to unvisited
			log.Info().Int("leased_browsers", b.browsers.Leased()).Ints64("leased_browsers", b.getLeased()).Msg("state monitor ping")
		case <-b.mainContext.Ctx.Done():
			log.Info().Msg("scan finished due to context complete")
			return
		case nav := <-b.navCh:
			log.Info().Int("leased_browsers", b.browsers.Leased()).Msg("processing nav")
			go b.crawl(nav)
			log.Info().Msg("Crawler to complete")
		case nav := <-b.attackCh:
			log.Info().Int("leased_browsers", b.browsers.Leased()).Msg("processing attack nav")
			go b.attack(nav)
		}
	}
}

func (b *Browserk) crawl(navs []*browserk.Navigation) {
	navCtx := b.mainContext.Copy()

	browser, port, err := b.browsers.Take(navCtx)
	if err != nil {
		log.Error().Err(err).Msg("failed to take browser")
		return
	}

	b.addLeased(browser.ID())
	defer b.removeLeased(browser.ID())

	crawler := crawler.New(b.cfg)
	if err := crawler.Init(); err != nil {
		b.browsers.Return(navCtx.Ctx, port)
		log.Error().Err(err).Msg("failed to init crawler")
		return
	}

	isFinal := false
	for i, nav := range navs {
		// we are on the last navigation of this path so we'll want to capture some stuff
		if i == len(navs)-1 {
			isFinal = true
		}

		ctx, cancel := context.WithTimeout(navCtx.Ctx, time.Second*45)
		navCtx.Ctx = ctx
		logger := log.With().
			Int64("browser_id", browser.ID()).
			Str("path", b.printActionStep(navs)).Int("step", i).
			Logger()
		navCtx.Log = &logger

		defer cancel()

		result, newNavs, err := crawler.Process(navCtx, browser, nav, isFinal)
		if err != nil {
			navCtx.Log.Error().Err(err).Msg("failed to process action")
			b.crawlGraph.FailNavigation(nav.ID)
			break
		}

		if isFinal {
			navCtx.Log.Info().Int("nav_count", len(newNavs)).Bool("is_final", isFinal).Msg("adding new navs")
			if err := b.crawlGraph.AddNavigations(newNavs); err != nil {
				navCtx.Log.Error().Err(err).Msg("failed to add new navigations")
			}
		}
		if err := b.crawlGraph.AddResult(result); err != nil {
			navCtx.Log.Error().Err(err).Msg("failed to add result")
		}
	}
	navCtx.Log.Info().Msg("closing browser")
	browser.Close()
	b.browsers.Return(navCtx.Ctx, port)
	b.readyCh <- struct{}{}
}

// attack iterates over the plugin, giving it it's own browser since we need to add
// the context specific to that plugin
func (b *Browserk) attack(navs []*browserk.NavigationWithResult) {

	navCtx := b.mainContext.Copy()

	browser, port, err := b.browsers.Take(navCtx)
	if err != nil {
		log.Error().Err(err).Msg("failed to take browser")
		return
	}
	logger := log.With().
		Int64("browser_id", browser.ID()).
		Logger()
	navCtx.Log = &logger
	b.addLeased(browser.ID())

	// TODO: Where to iterate over plugins? Keep in mind it'll need it's own navCtx for
	// potential hooks
	isFinal := false
	for i, nav := range navs {
		// we are on the last navigation of this path so we'll want to attack now
		if i == len(navs)-1 {
			isFinal = true
		}

		// Add GlobalHooks (stored xss function listener)

		if isFinal {

			// Create request iterator
			mIt := iterator.NewMessageIter(nav)
			for mIt.Rewind(); mIt.Valid(); mIt.Next() {
				// TODO: hash and store this for uniqueness otherwise we'll attack the same resources over
				// and over unnecessarily.
				req := mIt.Request()
				// TODO: Need to setup a matcher here so before ExecuteAction we can prepare the interception
				// alternatively, generate a new request from the browser and match that and replace everything...

				// Create injection iterator
				injIt := iterator.NewInjectionIter(req)
				for injIt.Rewind(); injIt.Valid(); injIt.Next() {
					injection, loc := injIt.Value()
					if loc == browserk.InjectQuery || loc == browserk.InjectFragment {
						k, _ := injIt.Key()
						v, _ := injIt.Value()
						navCtx.Log.Debug().Msgf("url: %s injection: name: %s value: %s", req.Request.Url, k, v)
					} else {
						navCtx.Log.Debug().Msgf("url: %s injection: %s", req.Request.Url, injection)
					}
				}
			}
		}

		ctx, cancel := context.WithTimeout(navCtx.Ctx, time.Second*45)
		browser.ExecuteAction(ctx, nav.Navigation)
		cancel()

	}
	navCtx.Log.Info().Msgf("closing attack browser %v", isFinal)
	browser.Close()
	b.browsers.Return(navCtx.Ctx, port)

	b.removeLeased(browser.ID())
	b.readyCh <- struct{}{}

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

	log.Info().Msg("Closing plugin store")
	err = b.pluginStore.Close()
	if err != nil {
		log.Warn().Err(err).Msg("failed to plugin store")
	}

	log.Info().Msg("Closing crawl graph")
	err = b.crawlGraph.Close()
	if err != nil {
		log.Warn().Err(err).Msg("failed to close crawlGraph")
	}
	return err
}

func (b *Browserk) printActionStep(navs []*browserk.Navigation) string {
	pathString := ""
	for i, path := range navs {
		if len(navs)-1 == i {
			pathString += fmt.Sprintf("%s %s", browserk.ActionTypeMap[path.Action.Type], path.Action)
			break
		}
		pathString += fmt.Sprintf("%s %s -> ", browserk.ActionTypeMap[path.Action.Type], path.Action)
	}
	return pathString
}
