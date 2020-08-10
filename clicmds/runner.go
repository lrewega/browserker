package clicmds

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/store"
)

// RunnerFlags configures how to run browserker
func RunnerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "url",
			Usage: "url as a start point",
			Value: "http://localhost/",
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "config to use",
			Value: "",
		},
		&cli.BoolFlag{
			Name:  "crawl",
			Usage: "only crawl, do not attack",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "browserktmp",
		},
		&cli.StringFlag{
			Name:  "report",
			Usage: "findings report json file",
			Value: "findings.json",
		},
		&cli.BoolFlag{
			Name:  "profile",
			Usage: "enable to profile cpu/mem",
			Value: false,
		},
		&cli.IntFlag{
			Name:  "numbrowsers",
			Usage: "max number of browsers to use in parallel",
			Value: 3,
		},
		&cli.IntFlag{
			Name:  "maxdepth",
			Usage: "max depth of nav paths to traverse",
			Value: 10,
		},
		&cli.StringFlag{
			Name:  "dot",
			Usage: "export crawl graph to DOT file",
			Value: "",
		},
		&cli.BoolFlag{
			Name:  "summary",
			Usage: "print summary of urls/graph actions taken",
			Value: true,
		},
	}
}

// Run browserker
func Run(cliCtx *cli.Context) error {
	if cliCtx.Bool("profile") {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	cfg := &browserk.Config{}
	cfg.FormData = &browserk.DefaultFormValues
	cfg.CrawlOnly = cliCtx.Bool("crawl")

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

	// TODO: remove this line after stabalizing, as it blows away all previous results
	os.RemoveAll(cfg.DataPath)

	crawl := store.NewCrawlGraph(cfg, cfg.DataPath+"/crawl")
	pluginStore := store.NewPluginStore(cfg.DataPath + "/plugin")
	browserk := scanner.New(cfg, crawl, pluginStore)
	log.Logger.Info().Msg("Starting browserker")

	scanContext := context.Background()
	if err := browserk.Init(scanContext); err != nil {
		log.Logger.Error().Err(err).Msg("failed to init engine")
		return err
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info().Msg("Ctrl-C Pressed, shutting down")
		err := browserk.Stop()
		log.Info().Msg("Giving a few seconds to sync db...")
		time.Sleep(time.Second * 10)
		if err != nil {
			log.Error().Err(err).Msg("failed to stop browserk")
		}
		os.Exit(1)
	}()

	start := time.Now()
	err := browserk.Start()
	if err != nil {
		log.Error().Err(err).Msg("browserk failure occurred")
	}

	if cliCtx.Bool("summary") {
		printSummary(crawl, cliCtx.String("dot"))
	}

	if cliCtx.String("report") != "" {
		writeReport(cliCtx.String("report"), cfg, crawl, pluginStore, start, time.Now())
	}

	return browserk.Stop()
}
