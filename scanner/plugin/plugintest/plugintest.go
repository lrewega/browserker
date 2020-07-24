package plugintest

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
	"gitlab.com/browserker/scanner/injections"
	"gitlab.com/browserker/scanner/iterator"
)

type AttackTests struct {
	FormHandler func(c *gin.Context)
	URL         string
}

func RunTestServer(testFilePath, formPath string, fn gin.HandlerFunc) (string, *http.Server) {
	router := gin.Default()
	router.Static("/forms", testFilePath)
	if fn != nil {
		router.Any(formPath, fn)
	}
	testListener, _ := net.Listen("tcp", ":0")
	_, testServerPort, _ := net.SplitHostPort(testListener.Addr().String())
	srv := &http.Server{
		Addr:    testListener.Addr().String(),
		Handler: router,
	}
	//testServerAddr := fmt.Sprintf("http://localhost:%s/", testServerPort)
	go func() {
		if err := srv.Serve(testListener); err != http.ErrServerClosed {
			log.Fatalf("Serve(): %s", err)
		}
	}()

	return testServerPort, srv
}

func GetNewNavPaths(bCtx *browserk.Context, pool *browser.GCDBrowserPool, target string) (browserk.Browser, string, []*browserk.NavigationWithResult, error) {

	b, port, err := pool.Take(bCtx)
	if err != nil {
		return nil, "", nil, err
	}

	crawl := crawler.New(&browserk.Config{})
	act := browserk.NewLoadURLAction(target)
	nav := browserk.NewNavigation(browserk.TrigCrawler, act)
	_, newNavs, err := crawl.Process(bCtx, b, nav, true)
	navResults := make([]*browserk.NavigationWithResult, 0)
	for _, newNav := range newNavs {
		r, _, _ := crawl.Process(bCtx, b, newNav, true)
		if r != nil {
			navResults = append(navResults, &browserk.NavigationWithResult{Navigation: newNav, Result: r})
		}

	}

	return b, port, navResults, err
}

func AttackWithPlugin(bCtx *browserk.Context, browser browserk.Browser, navResults []*browserk.NavigationWithResult) {
	for _, nav := range navResults {
		mIt := iterator.NewMessageIter(nav)
		for mIt.Rewind(); mIt.Valid(); mIt.Next() {
			req := mIt.Request()
			if req == nil || req.Request == nil {
				continue
			}
			// Create injection iterator
			injIt := iterator.NewInjectionIter(req)
			injector := injections.New(bCtx, browser, nav, mIt, injIt)
			for injIt.Rewind(); injIt.Valid(); injIt.Next() {
				bCtx.Log.Info().
					Str("location", injIt.Expr().Loc().String()).
					Str("method", req.Request.Method).
					Str("url", req.Request.Url).
					Str("body", req.Request.PostData).
					Msgf("auditing this injection")
				bCtx.PluginServicer.Inject(bCtx, injector)
				injIt.Expr().Reset()
			}
		}
	}
}
