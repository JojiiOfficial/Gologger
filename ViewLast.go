package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mkideal/cli"
)

//fmt.Println("\n" + os.Args[0] + " -s \"" + parseTime(config.LastStart) + "\" -u \"" + parseTime(config.LastEnd) + "\"")

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
		fmt.Println(parseTime(config.LastStart), parseTime(config.LastEnd))
		pullLogs(config, argv, config.LastStart, config.LastEnd, false)
		return nil
	},
}
