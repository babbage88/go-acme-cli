package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v3"
)

func printDnsRecordsTable(records []cloudflare.DNSRecord) {
	var colorInt int32 = 97
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(tw, "\x1b[1;%dm%s\t%s\t%s\t%s\t%s\t%s\x1b[0m\n", colorInt, "ID", "Name", "Content", "Type", "CreatedOn", "ModifiedOn")
	fmt.Fprintf(tw, "\x1b[1;%dm--\t----\t-------\t----\t---------\t----------\t\x1b[0m\n", colorInt)
	for _, v := range records {
		switch v.Type {
		case "A":
			colorInt = int32(96)
		case "CNAME":
			colorInt = int32(92)
		default:
			colorInt = int32(97)
		}
		fmt.Fprintf(tw, "\x1b[1;%dm%s\t%s\t%s\t%s\t%s\t%s\x1b[0m\n", colorInt, v.ID, v.Name, v.Content, v.Type, pretty.DateTimeSting(v.CreatedOn), pretty.DateTimeSting(v.ModifiedOn))
	}
	tw.Flush()
}

func printDnsAndZoneIdTable(domain string, zoneId string) error {
	tw := tabwriter.NewWriter(os.Stdout, 5, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "DomainName\t\tZoneID")
	fmt.Fprintln(tw, "----------\t\t------")
	fmt.Fprintf(tw, "%s\t\t%s\n", domain, zoneId)
	err := tw.Flush()
	return err
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

func cfDnsCommandFlags() []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "domain-name",
			Aliases: []string{"n"},
			Value:   "trahan.dev",
			Sources: cli.EnvVars("CF_DOMAIN_NAME"),
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

func cfDnsUpdateFlags() []cli.Flag {
	flags := cfDnsCommandFlags()
	updateFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "record-id",
			Aliases: []string{"i"},
			Sources: cli.EnvVars("CF_REC_UPDATE_ID"),
			Usage:   "The ID for Record you want to update.",
		},
		&cli.StringFlag{
			Name:    "new-content",
			Aliases: []string{"c"},
			Value:   "",
			Usage:   "The new content or Value for the record.",
		},
	}
	flags = append(flags, updateFlags...)
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
	zoneId, err := GetCloudFlareZoneIdByDomainName(envfile, domain)
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
			Name:                  "zone",
			Version:               versionNumber,
			Aliases:               []string{"get-zoneid"},
			Authors:               cfDnsComandAuthors(),
			Flags:                 cfDnsCommandFlags(),
			EnableShellCompletion: true,
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
			Name:                  "list",
			Version:               versionNumber,
			Authors:               cfDnsComandAuthors(),
			Aliases:               []string{"list-records"},
			Flags:                 cfDnsCommandFlags(),
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					records, err := GetCloudflareDnsListByDomainName(cmd.String("env-file"), cmd.String("domain-name"))
					if err != nil {
						msg := fmt.Sprintf("Error retrieving DNS Records %s", err.Error())
						slog.Error(msg)
						return err
					}
					printDnsRecordsTable(records)
					return err
				}

				records, err := GetCloudflareDnsListByDomainName(cmd.Args().Get(0), cmd.Args().Get(1))
				if err != nil {
					msg := pretty.PrettyErrorLogString("Error retrieving DNS Records %s", err.Error())
					pretty.PrintError(msg)
				}
				printDnsRecordsTable(records)
				return err
			},
		},
		{
			Name:                  "update",
			Version:               versionNumber,
			Authors:               cfDnsComandAuthors(),
			Aliases:               []string{"set", "update-record"},
			Flags:                 cfDnsUpdateFlags(),
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					records, err := GetCloudflareDnsListByDomainName(cmd.String("env-file"), cmd.String("domain-name"))
					if err != nil {
						msg := fmt.Sprintf("Error retrieving DNS Records %s", err.Error())
						slog.Error(msg)
						return err
					}
					printDnsRecordsTable(records)
					return err
				}

				records, err := GetCloudflareDnsListByDomainName(cmd.Args().Get(0), cmd.Args().Get(1))
				if err != nil {
					msg := pretty.PrettyErrorLogString("Error retrieving DNS Records %s", err.Error())
					pretty.PrintError(msg)
				}
				printDnsRecordsTable(records)
				return err
			},
		},
		{
			Name:                  "delete",
			Version:               versionNumber,
			Authors:               cfDnsComandAuthors(),
			Aliases:               []string{"rm", "remove-record"},
			Flags:                 cfDnsUpdateFlags(),
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					records, err := GetCloudflareDnsListByDomainName(cmd.String("env-file"), cmd.String("domain-name"))
					if err != nil {
						msg := fmt.Sprintf("Error retrieving DNS Records %s", err.Error())
						slog.Error(msg)
						return err
					}
					printDnsRecordsTable(records)
					return err
				}

				records, err := GetCloudflareDnsListByDomainName(cmd.Args().Get(0), cmd.Args().Get(1))
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
