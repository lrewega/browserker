package clicmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/store"
)

func DBViewFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "browserktmp",
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "config to use",
			Value: "",
		},
		&cli.BoolFlag{
			Name:  "navs",
			Usage: "prints navs",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "dot",
			Usage: "prints navs to file",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "out",
			Usage: "path to output directory, must specify --type",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "dom, nav, messages, requests, responses",
			Value: "",
		},
	}
}

func DBView(cliCtx *cli.Context) error {
	cfg := &browserk.Config{}
	cfg.FormData = &browserk.DefaultFormValues

	if cliCtx.String("config") == "" {
		cfg = &browserk.Config{
			URL:         cliCtx.String("url"),
			NumBrowsers: cliCtx.Int("numbrowsers"),
			MaxDepth:    cliCtx.Int("maxdepth"),
		}
	} else {
		data, err := ioutil.ReadFile(cliCtx.String("config"))
		if err != nil {
			return err
		}

		if err := toml.NewDecoder(strings.NewReader(string(data))).Decode(cfg); err != nil {
			return err
		}

		if cfg.URL == "" && cliCtx.String("url") != "" {
			cfg.URL = cliCtx.String("url")
		}
		if cfg.DataPath == "" && cliCtx.String("datadir") != "" {
			cfg.DataPath = cliCtx.String("datadir")
		}
	}

	crawl := store.NewCrawlGraph(cfg, cfg.DataPath+"/crawl")
	if err := crawl.Init(); err != nil {
		log.Error().Err(err).Msg("failed to init database for viewing")
		return err
	}

	if cliCtx.Bool("navs") {
		printSummary(crawl, cliCtx.String("dot"))
	}

	var err error
	dumpType := cliCtx.String("type")
	outDir := cliCtx.String("out")
	if outDir != "" && dumpType != "" {
		if err = dumpDataToPath(crawl, outDir, dumpType); err != nil {
			log.Error().Err(err).Msg("error dumping DOM files")
		}

	}

	log.Info().Msg("Closing db & syncing, please wait")
	err = crawl.Close()
	time.Sleep(5 * time.Second)
	return err
}

func dumpDataToPath(crawl *store.CrawlGraph, outDir, outType string) error {
	if err := os.MkdirAll(outDir, 0744); err != nil {
		return err
	}

	outType = strings.ToLower(outType)
	if outType == "nav" || outType == "navs" {
		saveNavs(crawl, outDir)
	}
	results, err := crawl.GetNavigationResults()
	if err != nil {
		return err
	}

	if results == nil {
		return fmt.Errorf("No result entries found")
	}

	fmt.Printf("Had %d results\n", len(results))

	for _, entry := range results {
		switch strings.ToLower(outType) {
		case "dom":
			saveDOM(entry, outDir)
		case "message", "messages":
			saveMessage(entry, outDir)
		case "request", "requests":
			saveRequest(entry, outDir)
		case "response", "responses":
			saveResponse(entry, outDir)
		}
	}
	return nil
}

func saveDOM(entry *browserk.NavigationResult, outDir string) {
	if entry.DOM == "" {
		return
	}
	fname := fmt.Sprintf("%x.html", entry.ID)
	contents := fmt.Sprintf("<!-- %s , %s -->\n%s", entry.StartURL, entry.EndURL, entry.DOM)
	err := ioutil.WriteFile(outDir+string(os.PathSeparator)+fname, []byte(contents), 0744)
	if err != nil {
		log.Error().Err(err).Str("url", entry.StartURL).Msg("failed to write file for url")
	}
}

func saveMessage(entry *browserk.NavigationResult, outDir string) {
	if entry.Messages == nil {
		return
	}
	type messages struct {
		StartURL string                  `json:"start_url"`
		EndURL   string                  `json:"end_url"`
		Messages []*browserk.HTTPMessage `json:"messages"`
	}
	m := &messages{StartURL: entry.StartURL, EndURL: entry.EndURL, Messages: entry.Messages}
	contents, err := json.Marshal(m)
	fname := fmt.Sprintf("%x-messages.json", entry.ID)
	err = ioutil.WriteFile(outDir+string(os.PathSeparator)+fname, []byte(contents), 0744)
	if err != nil {
		log.Error().Err(err).Msg("failed to write file for messages")
	}
}

func saveRequest(entry *browserk.NavigationResult, outDir string) {
	if entry.Messages == nil {
		return
	}
	type requests struct {
		StartURL string                  `json:"start_url"`
		EndURL   string                  `json:"end_url"`
		Requests []*browserk.HTTPRequest `json:"requests"`
	}
	reqs := make([]*browserk.HTTPRequest, 0)
	for _, m := range entry.Messages {
		reqs = append(reqs, m.Request)
	}
	m := &requests{StartURL: entry.StartURL, EndURL: entry.EndURL, Requests: reqs}
	contents, err := json.Marshal(m)
	fname := fmt.Sprintf("%x-requests.json", entry.ID)
	err = ioutil.WriteFile(outDir+string(os.PathSeparator)+fname, []byte(contents), 0744)
	if err != nil {
		log.Error().Err(err).Msg("failed to write file for requests")
	}
}

func saveResponse(entry *browserk.NavigationResult, outDir string) {
	if entry.Messages == nil {
		return
	}
	type responses struct {
		StartURL  string                   `json:"start_url"`
		EndURL    string                   `json:"end_url"`
		Responses []*browserk.HTTPResponse `json:"responses"`
	}
	resps := make([]*browserk.HTTPResponse, 0)
	for _, m := range entry.Messages {
		resps = append(resps, m.Response)
	}
	m := &responses{StartURL: entry.StartURL, EndURL: entry.EndURL, Responses: resps}
	contents, err := json.Marshal(m)
	fname := fmt.Sprintf("%x-responses.json", entry.ID)
	err = ioutil.WriteFile(outDir+string(os.PathSeparator)+fname, []byte(contents), 0744)
	if err != nil {
		log.Error().Err(err).Msg("failed to write file for responses")
	}
}

func saveNavs(crawl *store.CrawlGraph, outDir string) {
	visitedEntries := crawl.Find(nil, browserk.NavVisited, browserk.NavVisited, 5000)
	saveNav(visitedEntries, outDir, "visited")

	unvisitedEntries := crawl.Find(nil, browserk.NavUnvisited, browserk.NavUnvisited, 5000)
	saveNav(unvisitedEntries, outDir, "unvisited")

	inProcessEntries := crawl.Find(nil, browserk.NavInProcess, browserk.NavInProcess, 5000)
	saveNav(inProcessEntries, outDir, "inprocess")

	failedEntries := crawl.Find(nil, browserk.NavFailed, browserk.NavFailed, 5000)
	saveNav(failedEntries, outDir, "failed")

	auditedEntries := crawl.Find(nil, browserk.NavAudited, browserk.NavAudited, 5000)
	saveNav(auditedEntries, outDir, "audited")
}

func saveNav(navs [][]*browserk.Navigation, outDir, navType string) {
	ids := make(map[string]struct{})
	for _, nav := range navs {
		for _, entry := range nav {
			if entry == nil {
				continue
			}

			if _, exist := ids[string(entry.ID)]; !exist {
				ids[string(entry.ID)] = struct{}{}
			} else {
				continue
			}

			contents, err := json.Marshal(entry)
			fname := fmt.Sprintf("%x-%s.json", entry.ID, navType)
			err = ioutil.WriteFile(outDir+string(os.PathSeparator)+fname, []byte(contents), 0744)
			if err != nil {
				log.Error().Err(err).Msg("failed to write file for url")
			}
		}
	}
}
