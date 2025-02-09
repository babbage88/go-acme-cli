package main

import (
	"os"
	"time"

	"github.com/babbage88/go-acme-cli/internal/pretty"
)

func main() {

	logger := pretty.NewCustomLogger(os.Stdout, "INFO", int8(2), "|", true)
	logger.Info("Test Message")
	time.Sleep(time.Duration(time.Second * 2))
	logger.Debug("Test Debug message")
	time.Sleep(time.Duration(time.Second * 2))
	logger.Error("Error log test")

	/*
		app := commands.GetDnsRecords()
		if err := app.Run(os.Args); err != nil {
			log.Fatal(err)
		}
	*/
}
