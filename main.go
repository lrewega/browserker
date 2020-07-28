package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/clicmds"
)

func main() {
	app := cli.NewApp()
	app.Name = "Browserker Web App Scanner"
	app.Version = "0.1"
	app.Authors = []*cli.Author{{Name: "isaac dawson", Email: "isaac.dawson@gmail.com"}}
	app.Usage = "Analyzes a Web Site for Vulnerabilities"
	app.Commands = []*cli.Command{
		/*
			Enable when auth is ready
				{
					Name:    "testauth",
					Aliases: []string{"ta"},
					Usage:   "test authentication",
					Action:  clicmds.TestAuth,
					Flags:   clicmds.TestAuthFlags(),
				},
		*/
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "run browserker",
			Action:  clicmds.Run,
			Flags:   clicmds.RunnerFlags(),
		},
		{
			Name:    "replay",
			Aliases: nil,
			Usage:   "replay a specific navigation path",
			Action:  clicmds.ReplayNav,
			Flags:   clicmds.ReplayNavFlags(),
		},
		{
			Name:    "db",
			Aliases: nil,
			Usage:   "db viewer",
			Action:  clicmds.DBView,
			Flags:   clicmds.DBViewFlags(),
		},
	}
	fmt.Println(os.Args)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
