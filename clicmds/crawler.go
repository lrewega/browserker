package clicmds

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/emicklei/dot"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/store"
)

func CrawlerFlags() []cli.Flag {
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

// Crawler runs browserker crawler
func Crawler(cliCtx *cli.Context) error {
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

	err := browserk.Start()
	if err != nil {
		log.Error().Err(err).Msg("browserk failure occurred")
	}

	if cliCtx.Bool("summary") {
		printSummary(crawl, cliCtx.String("dot"))
	}

	return browserk.Stop()
}

func printSummary(crawl *store.CrawlGraph, dotFile string) error {
	results, err := crawl.GetNavigationResults()
	if err != nil {
		return err
	}

	if results == nil {
		return fmt.Errorf("No result entries found")
	}
	fmt.Printf("Had %d results\n", len(results))
	for _, entry := range results {
		if entry.Messages != nil {
			for _, m := range entry.Messages {
				if m.Request == nil {
					continue
				}
				fmt.Printf("URL visited: (DOC %s) %s\n", m.Request.DocumentURL, m.Request.Request.Url+m.Request.Request.UrlFragment)
			}
		}
	}

	visitedEntries := crawl.Find(nil, browserk.NavVisited, browserk.NavVisited, 999)
	printEntries(visitedEntries, "visited")

	unvisitedEntries := crawl.Find(nil, browserk.NavUnvisited, browserk.NavUnvisited, 999)
	printEntries(unvisitedEntries, "unvisited")
	inProcessEntries := crawl.Find(nil, browserk.NavInProcess, browserk.NavInProcess, 999)
	printEntries(inProcessEntries, "in process")
	failedEntries := crawl.Find(nil, browserk.NavFailed, browserk.NavFailed, 999)
	printEntries(failedEntries, "nav failed")
	if dotFile != "" {
		printDOT(dotFile, visitedEntries, unvisitedEntries, inProcessEntries, failedEntries)
	}
	return nil
}

func printEntries(entries [][]*browserk.Navigation, navType string) {
	fmt.Printf("Had %d %s entries\n", len(entries), navType)
	for _, paths := range entries {
		fmt.Printf("\n%s Path: \n", navType)
		for i, path := range paths {
			if len(paths)-1 == i {
				fmt.Printf("ID: %x %s", string(path.ID), path)
				break
			}
			fmt.Printf("ID: %x %s -> ", string(path.ID), path)
		}
	}
}

func printDOT(fileName string, visited, unvisited, inprocess, failed [][]*browserk.Navigation) {
	g := dot.NewGraph(dot.Directed)
	g.Attr("rankdir", "LR")
	subGraph(g.Subgraph("Visited"), visited)
	subGraph(g.Subgraph("Unvisited"), unvisited)
	subGraph(g.Subgraph("In Process"), inprocess)
	subGraph(g.Subgraph("Failed"), failed)

	ioutil.WriteFile(fileName, []byte(g.String()), 0677)
}

func subGraph(g *dot.Graph, entries [][]*browserk.Navigation) {
	for _, path := range entries {
		prev := g.Node(path[0].String())
		for i := 1; i < len(path); i++ {
			current := g.Node(path[i].String())
			g.Edge(prev, current)
			prev = current
		}
	}
}
