package store_test

import (
	"bytes"
	"os"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/store"
)

var testConfig = mock.MakeMockConfig()

func TestCrawlNavExists(t *testing.T) {
	os.RemoveAll("testdata/crawl")
	g := store.NewCrawlGraph(testConfig, "testdata/crawl")
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	nav := mock.MakeMockNavi([]byte{0, 1, 2})
	nav.OriginID = []byte{}
	if err := g.AddNavigation(nav); err != nil {
		t.Fatalf("error adding: %s\n", err)
	}
	if !g.NavExists(nav) {
		t.Fatalf("nav should have existed")
	}

	nonExist := mock.MakeMockNavi([]byte{0, 2, 2})
	if g.NavExists(nonExist) {
		t.Fatalf("nav should NOT existed")
	}
}

func TestCrawlOpenClose(t *testing.T) {
	os.RemoveAll("testdata/oc")
	g := store.NewCrawlGraph(testConfig, "testdata/oc")
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()
	for i := 1; i < 6; i++ {
		nav := mock.MakeMockNavi([]byte{0, byte(i), 2})
		nav.OriginID = []byte{0, byte(i - 1), 2}
		nav.Distance = i - 1

		if i == 1 {
			nav.OriginID = []byte{} // signals root
		}

		if err := g.AddNavigation(nav); err != nil {
			t.Fatalf("error adding: %s\n", err)
		}
	}
	testGetNavResults(t, g)
	g.Close()
	t.Log("REOPENING DB")
	g = store.NewCrawlGraph(testConfig, "testdata/oc")
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	// reset from inprocess -> unvisited
	limit := 5
	entries := g.Find(nil, browserk.NavInProcess, browserk.NavUnvisited, int64(limit))
	if len(entries) != limit {
		t.Fatalf("entries did not match limit got %d\n", len(entries))
	}

	testGetNavResults(t, g)
}

func TestCrawlGraph(t *testing.T) {
	os.RemoveAll("testdata/crawl")
	g := store.NewCrawlGraph(testConfig, "testdata/crawl")
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	nav := mock.MakeMockNavi([]byte{0, 1, 2})
	nav.OriginID = []byte{}

	if err := g.AddNavigation(nav); err != nil {
		t.Fatalf("error adding: %s\n", err)
	}

	result, err := g.GetNavigation(nav.ID)
	if err != nil {
		t.Fatalf("error reading back navigation: %s\n", err)
	}

	if result.Action == nil {
		t.Fatalf("action was nil")
	}

	if string(nav.ID) != string(result.ID) {
		t.Fatalf("%v != %v\n", nav.ID, result.ID)
	}

	if nav.Action.Type != result.Action.Type {
		t.Fatalf("%v != %v\n", nav.Action.Type, result.Action.Type)
	}
	_ = g.Find(nil, browserk.NavUnvisited, browserk.NavInProcess, 5)
}

func TestCrawlAddMultiple(t *testing.T) {
	path := "testdata/multi/crawl"
	os.RemoveAll(path)

	g := store.NewCrawlGraph(testConfig, path)
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	for i := 1; i < 11; i++ {
		nav := mock.MakeMockNavi([]byte{0, byte(i), 2})
		nav.OriginID = []byte{0, byte(i - 1), 2}
		nav.Distance = i - 1

		if i == 1 {
			nav.OriginID = []byte{} // signals root
		}

		if err := g.AddNavigation(nav); err != nil {
			t.Fatalf("error adding: %s\n", err)
		}
	}
	testGetNavResults(t, g)
}

func TestCrawlAddNavigations(t *testing.T) {
	path := "testdata/navis/crawl"
	os.RemoveAll(path)

	g := store.NewCrawlGraph(testConfig, path)
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	navs := make([]*browserk.Navigation, 0)
	for i := 1; i < 11; i++ {
		nav := mock.MakeMockNavi([]byte{0, byte(i), 2})
		nav.OriginID = []byte{0, byte(i - 1), 2}
		nav.Distance = i - 1

		if i == 1 {
			nav.OriginID = []byte{} // signals root
		}
		navs = append(navs, nav)
	}

	if err := g.AddNavigations(navs); err != nil {
		t.Fatalf("error calling  add navigations: %s\n", err)
	}

	testGetNavResults(t, g)
}

func TestCrawlAddNavigationsWithMax(t *testing.T) {
	path := "testdata/navis/crawl"
	os.RemoveAll(path)
	cfg := mock.MakeMockConfig()
	cfg.MaxActions = 5
	g := store.NewCrawlGraph(cfg, path)
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	navs := make([]*browserk.Navigation, 0)
	for i := 1; i < 11; i++ {
		nav := mock.MakeMockNavi([]byte{0, byte(i), 2})
		nav.OriginID = []byte{0, byte(i - 1), 2}
		nav.Distance = i - 1

		if i == 1 {
			nav.OriginID = []byte{} // signals root
		}
		navs = append(navs, nav)
	}

	if err := g.AddNavigations(navs); err != nil {
		t.Fatalf("error calling  add navigations: %s\n", err)
	}

	limit := 10
	entries := g.Find(nil, browserk.NavUnvisited, browserk.NavUnvisited, int64(limit))
	if len(entries) != cfg.MaxActions {
		t.Fatalf("entries did not match matx got %d\n", len(entries))
	}

}
func testGetNavResults(t *testing.T, g browserk.CrawlGrapher) {
	limit := 5
	entries := g.Find(nil, browserk.NavUnvisited, browserk.NavUnvisited, int64(limit))
	if len(entries) != limit {
		t.Fatalf("entries did not match limit got %d\n", len(entries))
	}

	for i := 0; i < len(entries); i++ {
		t.Logf("%d %d", i, len(entries[i]))
		if len(entries[i]) != i+1 {
			t.Fatalf("entry should have len %d got %d\n", i+1, len(entries[i]))
		}

		last := entries[i][len(entries[i])-1]
		if last.Distance != i {
			t.Fatalf("last entry distance should be %d got %d\n", i, last.Distance)
		}
	}

	entries = g.Find(nil, browserk.NavUnvisited, browserk.NavInProcess, int64(limit))
	if len(entries) != limit {
		t.Fatalf("entries did not match limit got %d\n", len(entries))
	}

	for i := 0; i < len(entries); i++ {
		t.Logf("%d %d", i, len(entries[i]))
		if entries[i][0].State != browserk.NavInProcess {
			t.Fatalf("expected in process state was %v\n", entries[i][0].State)
		}
	}
}

func TestCrawlAddResults(t *testing.T) {
	path := "testdata/results/crawl"
	os.RemoveAll(path)

	g := store.NewCrawlGraph(testConfig, path)
	if err := g.Init(); err != nil {
		t.Fatalf("error init graph: %s\n", err)
	}
	defer g.Close()

	for i := 1; i < 3; i++ {
		navResult := mock.MakeMockResult([]byte{0, byte(i), 2})

		if err := g.AddResult(navResult); err != nil {
			t.Fatalf("error adding: %s\n", err)
		}
	}

	res, err := g.GetNavigationResult([]byte{0, 1, 2})
	if err != nil {
		t.Fatalf("error getting nav result: %s\n", err)
	}
	if res.DOM != "<html>nav result</html>" {
		t.Fatalf("expected %s got [%s]", "<html>nav result</html>", res.DOM)
	}
	expectedBody := []byte("this is a body 0")

	if bytes.Compare(expectedBody, res.Messages[0].Response.Body) != 0 {
		t.Fatalf("expected body %s got %s", expectedBody, res.Messages[0].Response.Body)
	}
}
