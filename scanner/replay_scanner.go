package scanner

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/auth"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
	"gitlab.com/browserker/scanner/report"
)

// Browserk is our engine
type Replayer struct {
	cfg         *browserk.Config
	pluginStore browserk.PluginStorer
	crawlGraph  browserk.CrawlGrapher
	reporter    browserk.Reporter
	browsers    browserk.BrowserPool
	formHandler browserk.FormHandler

	mainContext *browserk.Context
	navID       []byte
}

// NewReplayer engine
func NewReplayer(cfg *browserk.Config, crawl browserk.CrawlGrapher, navID []byte) *Replayer {
	return &Replayer{
		cfg:         cfg,
		pluginStore: mock.MakeMockPluginStore(),
		crawlGraph:  crawl,
		navID:       navID,
		reporter:    report.New(crawl, mock.MakeMockPluginStore()),
	}
}

// Init the browsers and stores
func (b *Replayer) Init(ctx context.Context) error {
	target, err := url.Parse(b.cfg.URL)
	if err != nil {
		return err
	}
	cancelCtx, cancelFn := context.WithCancel(ctx)

	b.mainContext = browserk.NewContext(cancelCtx, cancelFn)

	pluginService := mock.MakeMockPluginServicer()

	b.mainContext.Auth = auth.New(b.cfg)
	b.mainContext.Scope = b.scopeService(target)
	b.mainContext.FormHandler = crawler.NewCrawlerFormHandler(b.cfg.FormData)
	b.mainContext.Reporter = b.reporter
	b.mainContext.Crawl = b.crawlGraph
	b.mainContext.PluginServicer = pluginService

	log.Info().Int("num_browsers", 1).Int("max_depth", b.cfg.MaxDepth).Msg("Initializing...")

	b.formHandler = crawler.NewCrawlerFormHandler(b.cfg.FormData)

	log.Logger.Info().Msg("starting leaser")
	leaser := browser.NewLocalLeaser()
	log.Logger.Info().Msg("leaser started")
	pool := browser.NewGCDBrowserPool(1, leaser)
	b.browsers = pool
	log.Logger.Info().Msg("starting browser pool")
	return pool.Init()
}

func (b *Replayer) scopeService(target *url.URL) browserk.ScopeService {
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
func (b *Replayer) Start() error {
	navs := b.crawlGraph.FindPathByNavID(b.mainContext.Ctx, b.navID)
	log.Info().Int("entries", len(navs)).Msg("Found entries")

	navCtx := b.mainContext.Copy()

	browser, port, err := b.browsers.Take(navCtx)
	if err != nil {
		log.Error().Err(err).Msg("failed to take browser")
		return err
	}

	crawler := crawler.New(b.cfg)
	if err := crawler.Init(); err != nil {
		b.browsers.Return(navCtx.Ctx, port)
		log.Error().Err(err).Msg("failed to init crawler")
		return err
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
			Int("of", len(navs)).
			Logger()
		navCtx.Log = &logger

		defer cancel()

		_, newNavs, err := crawler.Process(navCtx, browser, nav, isFinal)
		if err != nil {
			navCtx.Log.Error().Err(err).Msg("failed to process action")
			break
		}

		if isFinal {
			navCtx.Log.Debug().Int("nav_count", len(newNavs)).Str("NEW_NAVS", b.printActionStep(newNavs)).Msg("to be added")
		}
	}
	navCtx.Log.Info().Msg("closing browser")
	browser.Close()
	b.browsers.Return(navCtx.Ctx, port)
	return nil
}

// Stop the browsers
func (b *Replayer) Stop() error {

	log.Info().Msg("Completing Ctx")
	b.mainContext.CtxComplete()

	log.Info().Msg("Stopping browsers")
	err := b.browsers.Shutdown()
	if err != nil {
		log.Warn().Err(err).Msg("failed to close browsers")
	}

	log.Info().Msg("Closing crawl graph")
	err = b.crawlGraph.Close()
	if err != nil {
		log.Warn().Err(err).Msg("failed to close crawlGraph")
	}
	return err
}

func (b *Replayer) printActionStep(navs []*browserk.Navigation) string {
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
