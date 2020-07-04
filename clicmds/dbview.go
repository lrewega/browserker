package clicmds

import (
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
	}
}

func DBView(ctx *cli.Context) error {
	cfg := &browserk.Config{MaxDepth: 100}
	crawl := store.NewCrawlGraph(cfg, ctx.String("datadir"))
	if err := crawl.Init(); err != nil {
		log.Error().Err(err).Msg("failed to init database for viewing")
		return err
	}

	printSummary(crawl)

	log.Info().Msg("Closing db & syncing, please wait")
	err := crawl.Close()
	time.Sleep(5 * time.Second)
	return err
}
