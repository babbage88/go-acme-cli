package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("error loading .env", slog.String("error", err.Error()))
	}
	client := cloudflare.NewClient(
		option.WithAPIKey(os.Getenv("CLOUDFLARE_API_KEY")), // defaults to os.LookupEnv("CLOUDFLARE_API_KEY")
		option.WithAPIEmail(os.Getenv("CLOUDFLARE_EMAIL")), // defaults to os.LookupEnv("CLOUDFLARE_EMAIL")
	)

	recordResponse, err := client.DNS.Records.Get(
		context.TODO(),
		"023e105f4ecef8ad9ca31a8372d0c353",
		dns.RecordGetParams{
			ZoneID: cloudflare.F("023e105f4ecef8ad9ca31a8372d0c353"),
		},
	)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%+v\n", recordResponse)
}
