package clicmds

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/store"
)

func ReplayNavFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "browserktmp",
		},
		&cli.BoolFlag{
			Name:  "profile",
			Usage: "enable to profile cpu/mem",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "config to use",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "navID",
			Usage: "Navigation to replay for debugging",
		},
		&cli.IntFlag{
			Name:  "numbrowsers",
			Usage: "max number of browsers to use in parallel",
			Value: 1,
		},
		&cli.BoolFlag{
			Name:  "list",
			Usage: "list navs & navIDs",
			Value: false,
		},
	}
}

// ReplayNav reruns a specific nav
func ReplayNav(cliCtx *cli.Context) error {
	if cliCtx.Bool("profile") {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

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
	spew.Dump(cfg)

	if cliCtx.Bool("list") {
		printSummary(crawl)
		return nil
	}
	navID, err := hex.DecodeString(cliCtx.String("navID"))
	if err != nil {
		return err
	}

	replayer := scanner.NewReplayer(cfg, crawl, navID)
	if err := replayer.Init(context.Background()); err != nil {
		log.Logger.Error().Err(err).Msg("failed to init engine")
		return err
	}

	if err := replayer.Start(); err != nil {
		log.Error().Err(err).Msg("replayer failure occurred")
		return err
	}

	return replayer.Stop()
}
