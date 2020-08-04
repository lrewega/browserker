package crawler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
)

type crawlerTests struct {
	formHandler func(c *gin.Context)
	url         string
}

var leaser = browser.NewLocalLeaser()

func init() {
	//leaser.SetHeadless()
}

func testServer(path string, fn gin.HandlerFunc) (string, *http.Server) {
	router := gin.Default()
	router.Static("/forms", "testdata/forms")
	if fn != nil {
		router.Any(path, fn)
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

func TestCrawler(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()
	ctx := context.Background()

	called := false

	simpleCallFunc := func(c *gin.Context) {
		called = true
		resp := "<html><body>You made it!</body></html>"
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(resp))
	}

	type respData struct {
		Fname string `json:"fname"`
		Lname string `json:"lname"`
		Car   string `json:"cars"`
	}

	toTest := [...]crawlerTests{
		{
			simpleCallFunc,
			"http://localhost:%s/forms/target.html",
		},
		{
			func(c *gin.Context) {
				buf, err := ioutil.ReadAll(c.Request.Body)
				if err != nil {
					t.Logf("error reading body: %s\n", err)
					return
				}
				c.Request.Body = ioutil.NopCloser(bytes.NewReader(buf))

				dest := &respData{}
				if err := json.Unmarshal(buf, dest); err != nil {
					t.Logf("error Unmarshal body: %s\n", err)
					return
				}

				if dest.Fname == "Test" && dest.Lname == "User" && dest.Car == "volvo" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/floatingform.html",
		},
		{
			func(c *gin.Context) {
				fname, _ := c.GetQuery("fname")
				lname, _ := c.GetQuery("lname")

				if fname == "Test" && lname == "User" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/floatingformrealform.html",
		},

		{
			simpleCallFunc,
			"http://localhost:%s/forms/textclick.html",
		},
		{
			func(c *gin.Context) {
				fname, _ := c.GetQuery("fname")
				lname, _ := c.GetQuery("lname")

				if fname == "Test" && lname == "User" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/",
		},
		{
			func(c *gin.Context) {
				fname, _ := c.GetQuery("fname")
				lname, _ := c.GetQuery("lname")
				car, _ := c.GetQuery("cars")

				if fname == "Test" && lname == "User" && car == "volvo" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/select.html",
		},
		{
			func(c *gin.Context) {
				fname, _ := c.GetQuery("fname")
				lname, _ := c.GetQuery("lname")
				rad, _ := c.GetQuery("rad")

				if fname == "Test" && lname == "User" && rad == "rad1" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/radio.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseclick.html",
		},

		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmousedown.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseenter.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseleave.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseout.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseup.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/keydown.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/keypress.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/keyup.html",
		},

		{
			func(c *gin.Context) {
				username, _ := c.GetPostForm("username")
				password, _ := c.GetPostForm("password")
				if username == "testuser" && password == "testP@assw0rd1" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/login1.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmousedblclick.html",
		},
	}

	for _, crawlTest := range toTest {
		p, srv := testServer("/result/formResult", crawlTest.formHandler)
		defer srv.Shutdown(ctx)

		target := fmt.Sprintf(crawlTest.url, p)
		targetURL, _ := url.Parse(target)
		bCtx := mock.MakeMockContext(ctx, targetURL)
		bCtx.FormHandler = crawler.NewCrawlerFormHandler(&browserk.DefaultFormValues)
		bCtx.Scope = scanner.NewScopeService(targetURL)

		b, port, err := pool.Take(bCtx)
		if err != nil {
			t.Fatalf("error taking browser: %s\n", err)
		}

		crawl := crawler.New(&browserk.Config{})
		t.Logf("going to %s\n", target)
		act := browserk.NewLoadURLAction(target)
		nav := browserk.NewNavigation(browserk.TrigCrawler, act)
		_, newNavs, err := crawl.Process(bCtx, b, nav, true)
		if err != nil {
			t.Fatalf("error getting url %s\n", err)
		}

		if len(newNavs) == 0 {
			t.Fatal("did not find form nav action")
		}

		spew.Dump(newNavs)
		_, _, err = crawl.Process(bCtx, b, newNavs[0], true)
		if err != nil {
			t.Fatalf("failed to submit form %s\n", err)
		}

		if !called {
			t.Fatalf("form was not submitted: %s\n", target)
		}
		called = false
		pool.Return(ctx, port)
		srv.Shutdown(ctx)
	}

}
