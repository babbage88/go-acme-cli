package cloudflare

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func NewCloudflareClient() (*cloudflare.Client, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Error("error loading .env", slog.String("error", err.Error()))
	}

	client := cloudflare.NewClient(
		option.WithAPIKey(os.Getenv("CLOUDFLARE_API_KEY")), // defaults to os.LookupEnv("CLOUDFLARE_API_KEY")
		option.WithAPIEmail(os.Getenv("CLOUDFLARE_EMAIL")), // defaults to os.LookupEnv("CLOUDFLARE_EMAIL")
	)

	if client == nil {
		slog.Error("Error creating cloudflare.Client, recieved nil pointer. Verify CLOUDFLARE_API_KEY and CLOUDFLARE_EMAIL Env vars")
		return nil, fmt.Errorf("rror creating cloudflare.Client, recieved nil pointer. Verify CLOUDFLARE_API_KEY and CLOUDFLARE_EMAIL Env vars", os.Getenv("CLOUDFLARE_EMAIL"))
	}

	return client, nil
}

func getDnsRecordsList(zoneId string) error {
	client, err := NewCloudflareClient()
	if err != nil {
		slog.Error("Error creating cloudflare.Client, recieved nil pointer. Verify CLOUDFLARE_API_KEY and CLOUDFLARE_EMAIL Env vars")
		return err
	}
	page, err := client.DNS.Records.List(context.TODO(), dns.RecordListParams{
		ZoneID: cloudflare.F("023e105f4ecef8ad9ca31a8372d0c353"),
	})
	if err != nil {
		slog.Error("Error retrieving DNS records", slog.String("error", err.Error()))
		return err
	}

	records := make([]dns.RecordResponse, 0)

	records = append(records, page.Result...)
	page, err = page.GetNextPage()
	if err != nil {
		slog.Error("Error retieving next page of records.", slog.String("error", err.Error()))
		return err
	}
	for page != nil {
		records = append(records, page.Result...)
		page, err = page.GetNextPage()
	}
}

func GetDnsRecords(appInstance *cli.App) {
	appInst = &cli.App{
		Name:                 "infra-cli",
		Version:              "0.0.10",
		Compiled:             time.Now(),
		Args:                 true,
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Justin Trahan",
				Email: "justin@trahan.dev",
			},
		},
	}
}
