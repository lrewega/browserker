package clicmds_test

import (
	"testing"

	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/clicmds"
)

func TestCrawler(t *testing.T) {
	t.Skip()
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name:    "crawl",
			Aliases: []string{"c"},
			Usage:   "crawl only",
			Action:  clicmds.Crawler,
			Flags:   clicmds.CrawlerFlags(),
		},
	}
	err := app.Run([]string{"app", "c", "--url", "http://example.com", "--config", "../configs/wivet.toml"})
	if err != nil {
		t.Fatalf("err: %s\n", err)
	}
}
