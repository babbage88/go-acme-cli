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
	"github.com/urfave/cli/v3"
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
		slog.Error("Error initializing cf api client. Verify token.")
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
		slog.Error(msg)
		return zoneID, err
	}

	return zoneID, err
}

func printDnsRecordsTable(records []cloudflare.DNSRecord) {
	tw := tabwriter.NewWriter(os.Stdout, 5, 1, 2, '\t', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", "ID", "Name", "Content", "Type", "CreatedOn", "ModifiedOn")
	for _, v := range records {
		pretty.Print(fmt.Sprintf("DNSRecord %s\t%s\t%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Content, v.Type, pretty.DateTimeSting(v.CreatedOn), pretty.DateTimeSting(v.ModifiedOn)))
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Content, v.Type, pretty.DateTimeSting(v.CreatedOn), pretty.DateTimeSting(v.ModifiedOn))
	}
}

type UrFaveCliDocumentationSucks struct {
	Name  string `json:"authorName"`
	Email string `json:"email"`
}

func (author *UrFaveCliDocumentationSucks) String() string {
	return fmt.Sprintf("Name: %s Email: %s", author.Name, author.Email)
}

func cfDnsCommandFlags() []cli.Flag {
	flags := []cli.Flag{
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
	}
	return flags
}

func cfDnsComandAuthors() []any {
	authors := []any{
		&UrFaveCliDocumentationSucks{
			Name:  "Justin Trahan",
			Email: "justin@trahan.dev",
		},
	}
	return authors
}

func getZoneIdCmd(envfile string, domain string) error {
	zoneId, err := GetCloudFlareZoneIdForDomainName(envfile, domain)
	if err != nil {
		msg := fmt.Sprintf("Error retrieving DNS Records %s", err.Error())
		logger.Error(msg)
		return err
	}
	msg := fmt.Sprintf("Domain: %s ZoneId: %s", domain, zoneId)
	logger.Info(msg)
	return nil
}

func GetDnsSubCommands() []*cli.Command {
	dnsCmds := []*cli.Command{
		{
			Name:    "dns",
			Version: "1.0.0",
			Aliases: []string{"get-zoneid"},
			Authors: cfDnsComandAuthors(),
			Flags:   cfDnsCommandFlags(),
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					err := getZoneIdCmd(cmd.String("env-file"), cmd.String("domain-name"))
					if err != nil {
						return err
					}
					return nil
				}
				err = getZoneIdCmd(cmd.Args().Get(0), cmd.Args().Get(1))
				if err != nil {
					return err
				}
				return err
			},
		},
		{
			Name:    "list",
			Version: "1.0.0",
			Authors: cfDnsComandAuthors(),
			Flags:   cfDnsCommandFlags(),
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {

					records, err := getCloudflareDnsListByDomainName(cmd.String("env-file"), cmd.String("domain-name"))
					if err != nil {
						msg := fmt.Sprintf("Error retrieving DNS Records %s", err.Error())
						slog.Error(msg)
						return err
					}
					printDnsRecordsTable(records)
					return nil

				}
				log.Printf("args: %+v", cmd.Args())

				records, err := getCloudflareDnsListByDomainName(cmd.Args().Get(0), cmd.Args().Get(1))
				if err != nil {
					msg := pretty.PrettyErrorLogString("Error retrieving DNS Records %s", err.Error())
					pretty.PrintError(msg)
				}
				printDnsRecordsTable(records)
				return nil
			},
		},
	}
	return dnsCmds
}

func CoreInfraCommand() *cli.Command {
	cmd := &cli.Command{
		Name:     "goinfra",
		Version:  "v1.0.0",
		Authors:  cfDnsComandAuthors(),
		Flags:    cfDnsCommandFlags(),
		Commands: GetDnsSubCommands(),
	}
	return cmd
}
