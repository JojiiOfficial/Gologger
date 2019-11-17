package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

var help = cli.HelpCommand("Display help information")

func main() {
	if err := cli.Root(viewCMD,
		cli.Tree(viewLastCMD),
		cli.Tree(help),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
