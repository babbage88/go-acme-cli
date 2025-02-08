package main

import (
	"log"
	"os"

	"github.com/babbage88/go-acme-cli/commands"
)

func main() {
	app := commands.GetDnsRecords()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
