package main

import (
	"os"

	"github.com/babbage88/go-acme-cli/commands"
	"github.com/babbage88/go-acme-cli/internal/pretty"
)

func main() {
	var logger = pretty.NewCustomLogger(os.Stdout, "DEBUG", 1, "|", true)
	app := commands.GetDnsRecords()
	if err := app.Run(os.Args); err != nil {
		logger.Error(err.Error())
	}
}
