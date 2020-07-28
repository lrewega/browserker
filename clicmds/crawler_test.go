package clicmds_test

import (
	"testing"

	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/clicmds"
)

func TestRun(t *testing.T) {
	t.Skip()
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "run",
			Action:  clicmds.Run,
			Flags:   clicmds.RunnerFlags(),
		},
	}
	err := app.Run([]string{"app", "run", "--config", "../configs/dvwa_lowdepth.toml"})
	if err != nil {
		t.Fatalf("err: %s\n", err)
	}
}
