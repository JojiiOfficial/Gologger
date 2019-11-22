package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mkideal/cli"
)

var viewLastCMD = &cli.Command{
	Name:    "viewlast",
	Aliases: []string{"vl", "last", "viewlastcommand"},
	Desc:    "Show logs between the timestamps from last view",
	Argv:    func() interface{} { return new(viewT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*viewT)
		config, err := checkConfig(argv.ConfigFile)
		if err != nil {
			fmt.Println("Error creating config:", err.Error())
			return nil
		}
		if argv.NoColor || os.Getenv("NO_COLOR") == "true" {
			color.NoColor = true
		}
		if config.LastStart == config.LastEnd || (config.LastStart == 0 && config.LastEnd == 0) {
			return errors.New("No history")
		}
		start := config.LastStart
		end := config.LastEnd
		if len(argv.Until) > 0 {
			end, err = parseTimeParam(argv.Until)
		}
		if len(argv.Since) > 0 {
			start, err = parseTimeParam(argv.Since)
		}
		if err != nil {
			fmt.Println("Error parsing time: " + err.Error())
		}
		pullLogs(config, argv, start, end, false)
		return nil
	},
}
