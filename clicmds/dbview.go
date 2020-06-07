package clicmds

import (
	"fmt"
	"time"

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
		&cli.BoolFlag{
			Name:  "navs",
			Usage: "prints navs",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "urls",
			Usage: "prints urls",
			Value: false,
		},
	}
}

func DBView(ctx *cli.Context) error {
	cfg := &browserk.Config{MaxDepth: 100}
	crawl := store.NewCrawlGraph(cfg, ctx.String("datadir"))
	if err := crawl.Init(); err != nil {
		log.Error().Err(err).Msg("failed to init database for viewing")
		return err
	}

	if ctx.Bool("urls") {
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
					fmt.Printf("URL visited: (DOC %s) %s\n", m.Request.DocumentURL, m.Request.Request.Url)
				}
			}
		}
		entries := crawl.Find(nil, browserk.NavVisited, browserk.NavVisited, 999)
		fmt.Printf("Had %d entries\n", len(entries))

		for _, paths := range entries {
			fmt.Printf("Path: \n")
			for i, path := range paths {
				if len(paths)-1 == i {
					fmt.Printf("%s %s\n", browserk.ActionTypeMap[path.Action.Type], path.Action)
					break
				}
				fmt.Printf("%s %s -> ", browserk.ActionTypeMap[path.Action.Type], path.Action)

			}
		}
	}
	log.Info().Msg("Closing db & syncing, please wait")
	err := crawl.Close()
	time.Sleep(5 * time.Second)
	return err
}
