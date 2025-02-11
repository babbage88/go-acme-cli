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
	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var logger = pretty.NewCustomLogger(os.Stdout, "DEBUG", 1, "|", true)

type DnsRecord struct {
	Id       string    `json:"id"`
	ZoneId   string    `json:"zoneId"`
	Name     string    `json:"name"`
	Content  string    `json:"content"`
	Type     string    `json:"type"`
	Modified time.Time `json:"lastModified"`
	Created  time.Time `json:"created"`
}

func NewCloudflareAPIClient(envfile string) (*cloudflare.API, error) {
	err := godotenv.Load(envfile)
	if err != nil {
		slog.Error("error loading .env", slog.String("error", err.Error()))
	}

	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_DNS_API_TOKEN"))
	if err != nil {
		logger.Error("Error initializing cf api client. Verify token.")
		return api, err
	}

	return api, nil
}

func getCloudflareDnsListByDomainName(envfile string, domainName string) ([]cloudflare.DNSRecord, error) {
	records := make([]cloudflare.DNSRecord, 0)
	api, err := NewCloudflareAPIClient(envfile)
	if err != nil {
		return records, err
	}

	zoneID, err := api.ZoneIDByName(domainName)
	if err != nil {
		slog.Debug("Error retrieving ZoneId for domain name", slog.String("DomainName", domainName))
		return records, err
	}

	records, _, err = api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return records, err
	}
	return records, nil
}

func GetCloudFlareZoneIdForDomainName(envfile string, domainName string) (string, error) {
	api, err := NewCloudflareAPIClient(envfile)
	if err != nil {
		return "", err
	}

	zoneID, err := api.ZoneIDByName(domainName)
	if err != nil {
		msg := fmt.Sprintf("Error retrieving ZoneId for Domain: %s error: %s", domainName, err.Error())
		logger.Error(msg)
		return zoneID, err
	}

	return zoneID, err
}

/*
func getDnsRecordsList(envfile string, zoneId string) ([]dns.RecordResponse, error) {
	client, err := NewCloudflareClient(envfile)
	records := make([]dns.RecordResponse, 0)
	if err != nil {
		slog.Error("Error creating cloudflare.Client, recieved nil pointer. Verify CLOUDFLARE_API_KEY and CLOUDFLARE_EMAIL Env vars")
		return records, err
	}
	page, err := client.DNS.Records.List(context.TODO(), dns.RecordListParams{
		ZoneID: cloudflare.F(zoneId),
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
*/

func printDnsRecordsTable(records []cloudflare.DNSRecord) {
	tw := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', 0)
	for _, v := range records {
		logger.Info(fmt.Sprintf("Record ID %s", v.ID))
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
				Name:    "domain-name",
				Aliases: []string{"n"},
				Value:   "trahan.dev",
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
				records, err := getCloudflareDnsListByDomainName(cCtx.String("env-file"), cCtx.String("domain-name"))
				if err != nil {
					msg := fmt.Sprintf("Error retrieving DNS Records %s", err.Error())
					logger.Error(msg)
					return err
				}
				printDnsRecordsTable(records)
				return nil

			}
			log.Printf("args: %+v", cCtx.Args())

			records, err := getCloudflareDnsListByDomainName(cCtx.Args().Get(0), cCtx.Args().Get(1))
			if err != nil {
				msg := pretty.PrettyErrorLogString("Error retrieving DNS Records %s", err.Error())
				logger.Error(msg)
			}
			printDnsRecordsTable(records)
			return nil
		},
	}
	return appInst
}
