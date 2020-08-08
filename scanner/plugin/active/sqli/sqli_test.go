package sqli_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
	"gitlab.com/browserker/scanner/plugin/active/sqli"
	"gitlab.com/browserker/scanner/plugin/plugintest"

	_ "net/http/pprof"

	_ "net/http"
)

var leaser = browser.NewLocalLeaser()

func init() {
	leaser.SetHeadless()
	//leaser.SetProxy("http://127.0.0.1:8080")
}

func TestSQLi(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()
	ctx := context.Background()
	calledCh := make(chan struct{})

	toTest := [...]plugintest.AttackTests{
		{
			FormHandler: func(c *gin.Context) {
				user := c.PostForm("username")
				resp := "<html><body>You made it!</body></html>"
				if user == "'+(select(sleep(15)))+'" {
					t.Logf("calling sleep...")
					time.Sleep(time.Second * 15)
					calledCh <- struct{}{}
				}

				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			URL: "http://localhost:%s/forms/simplePOST.html",
		},

		{
			FormHandler: func(c *gin.Context) {
				user, _ := c.GetQuery("username")
				resp := "<html><body>You made it!</body></html>"
				if user == "'+(select(sleep(15)))+'" {
					t.Logf("calling sleep...")
					time.Sleep(time.Second * 2)
					calledCh <- struct{}{}
				}

				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			URL: "http://localhost:%s/forms/simpleGET.html",
		},
		{
			FormHandler: func(c *gin.Context) {
				user, _ := c.GetQuery("username")
				resp := "<html><body>You made it!</body></html>"
				if user == "'+(select(sleep(15)))+'" {
					t.Logf("calling sleep (for timeout)...")
					time.Sleep(time.Second * 2)
					calledCh <- struct{}{}
				}

				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			URL: "http://localhost:%s/forms/simpleGET.html#timeout",
		},
	}

	for _, attackTest := range toTest {
		p, srv := plugintest.RunTestServer("../../plugintest/testdata/forms", "/form/result", attackTest.FormHandler)
		defer srv.Shutdown(ctx)

		target := fmt.Sprintf(attackTest.URL, p)
		targetURL, _ := url.Parse(target)
		ctx := context.Background()
		bCtx := mock.MakeMockContext(ctx, targetURL)
		bCtx.FormHandler = crawler.NewCrawlerFormHandler(&browserk.DefaultFormValues)
		bCtx.Scope = scanner.NewScopeService(targetURL)

		browser, port, navResults, err := plugintest.GetNewNavPaths(bCtx, pool, target)
		if err != nil {
			t.Fatalf("error getting new nav paths: %s\n", err)
		}

		if len(navResults) == 0 {
			t.Fatal("did not find form nav action")
		}

		bCtx.PluginServicer = mock.MakeMockPluginServicer()
		bCtx.PluginServicer.Register(sqli.New(bCtx.PluginServicer))
		t.Logf("Attacking With Plugin")
		plugintest.AttackWithPlugin(bCtx, browser, navResults)
		callCnt := 0

		respTimer := time.NewTimer(time.Second * 2)
		defer respTimer.Stop()

		for i := 0; i < 2; i++ {
			select {
			case <-calledCh:
				t.Logf("got call from timing attack")
				callCnt++
			case <-respTimer.C:
				t.Logf("timed out waiting for call: iter: %d, callCnt: %d", i, callCnt)
				break
			}
		}
		if callCnt != 2 {
			t.Fatalf("attack was not successful: %s, callcnt %d\n", target, callCnt)
		}

		pool.Return(ctx, port)
		srv.Shutdown(ctx)
	}
}
