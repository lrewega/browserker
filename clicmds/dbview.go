package clicmds

import (
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
			Value: true,
		},
		&cli.StringFlag{
			Name:  "dot",
			Usage: "prints navs to file",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "dumpdom",
			Usage: "dumps serialized DOM for each navigation w/result to the specified directory",
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
	dumpDOM := cliCtx.String("dumpdom")
	if dumpDOM != "" {
		if err = dumpDOMToPath(crawl, dumpDOM); err != nil {
			log.Error().Err(err).Msg("error dumping DOM files")
		}
	}

	log.Info().Msg("Closing db & syncing, please wait")
	err = crawl.Close()
	time.Sleep(5 * time.Second)
	return err
}

func dumpDOMToPath(crawl *store.CrawlGraph, dumpDOM string) error {
	if err := os.MkdirAll(dumpDOM, 0744); err != nil {
		return err
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
		if entry.DOM == "" {
			continue
		}
		fname := fmt.Sprintf("%x.html", entry.ID)
		contents := fmt.Sprintf("<!-- %s , %s -->\n%s", entry.StartURL, entry.EndURL, entry.DOM)
		err := ioutil.WriteFile(dumpDOM+string(os.PathSeparator)+fname, []byte(contents), 0744)
		if err != nil {
			log.Error().Err(err).Str("url", entry.StartURL).Msg("failed to write file for url")
		}
	}
	return nil
}
