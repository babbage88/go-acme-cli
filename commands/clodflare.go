package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

const versionNumber = "v1.0.0"

var logger = pretty.NewCustomLogger(os.Stdout, "DEBUG", 1, "|", true)

type GoInfraCli struct {
	Name        string       `json:"name"`
	BaseCommand *cli.Command `json:"baseCmd"`
	Version     string       `json:"version"`
}

func (c *GoInfraCli) SubCommands() []*cli.Command {
	return c.BaseCommand.Commands
}

func (c *GoInfraCli) Authors(a *UrFaveCliDocumentationSucks) string {
	return a.String()
}

func (c *GoInfraCli) GetDnsSubCommands(sub *cli.Command) error {
	start := len(c.BaseCommand.Commands)
	c.BaseCommand.Commands = append(c.BaseCommand.Commands, sub)
	if len(c.BaseCommand.Commands) == start {
		err := fmt.Errorf("error adding SubCommands, length %d did not change", start)
		return err
	}
	return nil
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
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", "ID", "Name", "Content", "Type", "CreatedOn", "ModifiedOn")
	for _, v := range records {
		pretty.Print(fmt.Sprintf("DNSRecord %s\t%s\t%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Content, v.Type, pretty.DateTimeSting(v.CreatedOn), pretty.DateTimeSting(v.ModifiedOn)))
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Content, v.Type, pretty.DateTimeSting(v.CreatedOn), pretty.DateTimeSting(v.ModifiedOn))
	}
}

func printDnsAndZoneIdTable(domain string, zoneId string) error {
	tw := tabwriter.NewWriter(os.Stdout, 5, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "DomainName\t\tZoneID")
	fmt.Fprintf(tw, "%s\t\t%s\n", domain, zoneId)
	err := tw.Flush()
	return err
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
	printDnsAndZoneIdTable(domain, zoneId)

	return nil
}

func GetDnsSubCommands() []*cli.Command {
	dnsSubCmds := []*cli.Command{
		{
			Name:    "zone",
			Version: versionNumber,
			Aliases: []string{"get-zoneid"},
			Authors: cfDnsComandAuthors(),
			Flags:   cfDnsCommandFlags(),
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					err := getZoneIdCmd(cmd.String("env-file"), cmd.String("domain-name"))
					return err
				}
				err = getZoneIdCmd(cmd.Args().Get(0), cmd.Args().Get(1))
				return err
			},
		},
		{
			Name:    "list",
			Version: versionNumber,
			Authors: cfDnsComandAuthors(),
			Aliases: []string{"list-records"},
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
					return err
				}

				records, err := getCloudflareDnsListByDomainName(cmd.Args().Get(0), cmd.Args().Get(1))
				if err != nil {
					msg := pretty.PrettyErrorLogString("Error retrieving DNS Records %s", err.Error())
					pretty.PrintError(msg)
				}
				printDnsRecordsTable(records)
				return err
			},
		},
	}
	return dnsSubCmds
}

func DnsBaseCommand() []*cli.Command {
	cmd := []*cli.Command{
		{
			Name:                  "dns",
			EnableShellCompletion: true,
			Version:               versionNumber,
			Authors:               cfDnsComandAuthors(),
			Commands:              GetDnsSubCommands(),
		},
	}
	return cmd
}

func CoreInfraCommand() *cli.Command {
	cmd := &cli.Command{
		Name:                  "goinfra",
		EnableShellCompletion: true,
		Version:               "v1.0.0",
		Authors:               cfDnsComandAuthors(),
		// Flags:    cfDnsCommandFlags(),
		Commands: DnsBaseCommand(),
	}
	return cmd
}
