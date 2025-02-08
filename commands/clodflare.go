package commands

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"

	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func NewCloudflareClient(envfile string) (*cloudflare.Client, error) {
	err := godotenv.Load(envfile)
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

func getDnsRecordsList(envfile string, zoneId string) ([]dns.RecordResponse, error) {
	client, err := NewCloudflareClient(envfile)
	records := make([]dns.RecordResponse, 0)
	if err != nil {
		slog.Error("Error creating cloudflare.Client, recieved nil pointer. Verify CLOUDFLARE_API_KEY and CLOUDFLARE_EMAIL Env vars")
		return records, err
	}
	page, err := client.DNS.Records.List(context.TODO(), dns.RecordListParams{
		ZoneID: cloudflare.F("023e105f4ecef8ad9ca31a8372d0c353"),
	})
	if err != nil {
		slog.Error("Error retrieving DNS records", slog.String("error", err.Error()))
		return records, err
	}

	records = append(records, page.Result...)
	page, err = page.GetNextPage()
	if err != nil {
		slog.Error("Error retieving next page of records.", slog.String("error", err.Error()))
		return records, err
	}

	for page != nil {
		records = append(records, page.Result...)
		page, err = page.GetNextPage()
		if err != nil {
			slog.Error("Error retieving next page of records.", slog.String("error", err.Error()))
			return records, err
		}
	}

	return records, err
}

func printDnsRecordsTable(records []dns.RecordResponse) {
	tw := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', 0)
	for _, v := range records {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", "ID", "Name", "Content", "Data", "Type", "CreatedOn", "ModifiedOn")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Content, v.Data, v.Type, pretty.DateTimeSting(v.CreatedOn), pretty.DateTimeSting(v.ModifiedOn))
	}
}

func GetDnsRecords() (appInst *cli.App) {
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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "zone-id",
				Aliases: []string{"z"},
				Value:   "1a03a1886dc5855341b01d0afa9fa3c3",
				Usage:   "Cloudflare Zone Id to retrieve records for.",
			},
			&cli.StringFlag{
				Name:    "env-file",
				Aliases: []string{"e"},
				Value:   ".env",
				Usage:   ".env file to use to load Cloudflare API keys and Zone ID",
			},
		},
		Action: func(cCtx *cli.Context) (err error) {
			if cCtx.NArg() == 0 {
				records, err := getDnsRecordsList(cCtx.String("env-file"), cCtx.String("zone-id"))
				if err != nil {
					msg := pretty.PrettyErrorLogString("Error retrieving DNS Records %s", err.Error())
					log.Fatalf("%s", msg)
					return err
				}
				printDnsRecordsTable(records)
				return nil

			}
			log.Printf("args: %+v", cCtx.Args())

			records, err := getDnsRecordsList(cCtx.Args().Get(0), cCtx.Args().Get(1))
			if err != nil {
				msg := pretty.PrettyErrorLogString("Error retrieving DNS Records %s", err.Error())
				log.Fatalf("%s", msg)
			}
			printDnsRecordsTable(records)
			return nil
		},
	}
	return appInst
}
