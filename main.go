package main

import (
	"context"
	"os"

	"github.com/babbage88/go-acme-cli/commands"
	"github.com/babbage88/go-acme-cli/internal/pretty"
)

func main() {
	// Load Base go-infra-cli
	logger := pretty.NewCustomLogger(os.Stdout, "DEBUG", 1, "|", true)
	cmd := commands.CoreInfraCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logger.Error(err.Error())
	}
}
