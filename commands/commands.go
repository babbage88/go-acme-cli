package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v3"
)

func DnsBaseCommand() []*cli.Command {
	cmd := []*cli.Command{
		{
			Name:                  "dns",
			EnableShellCompletion: true,
			Version:               versionNumber,
			Authors:               cfDnsComandAuthors(),
			Commands:              GetDnsSubCommands(),
			Flags:                 cfDnsSubcommandFlags(),
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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "env-file",
				Aliases: []string{"e"},
				Value:   ".env",
				Usage:   ".env file to use to load Cloudflare API keys and Zone ID",
			},
		},
		Commands: DnsBaseCommand(),
	}
	return cmd
}

func cfDnsSubcommandFlags() []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "domain-name",
			Aliases: []string{"n"},
			Value:   "trahan.dev",
			Sources: cli.EnvVars("CF_DOMAIN_NAME"),
			Usage:   "Cloudflare Zone Id to retrieve records for.",
		},
		&cli.StringFlag{
			Name:    "new-content",
			Aliases: []string{"c"},
			Usage:   "The new content or Value for the record.",
		},
		&cli.StringFlag{
			Name:  "record-name",
			Usage: "The name for the dns record",
		},
		&cli.StringFlag{
			Name:    "type",
			Aliases: []string{"t", "record-type"},
			Usage:   "The type of dns record: A, AAAA, CNAME, MX, TXT",
		},
		&cli.UintFlag{
			Name:    "priority",
			Aliases: []string{"p", "record-priority"},
			Usage:   "DNS Record priority",
		},
		&cli.IntFlag{
			Name:    "ttl",
			Aliases: []string{"record-ttl"},
			Value:   3600,
			Usage:   "new ttl for recird",
		},
		&cli.BoolFlag{
			Name:  "proxied",
			Usage: "Whether the record is proxied via cloudflare",
		},
		&cli.StringFlag{
			Name:    "comment",
			Aliases: []string{"record-comment"},
			Usage:   "Comment for the dns record",
		},
		&cli.StringSliceFlag{
			Name:    "tags",
			Aliases: []string{"record-tags"},
			Usage:   "tags for record",
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
			Category:              "dns",
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
			Category:              "dns",
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
			Name:    "update",
			Version: versionNumber,
			Authors: cfDnsComandAuthors(),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "update-record-id",
					Required: true,
					Aliases:  []string{"record-id"},
					Sources:  cli.EnvVars("CF_REC_UPDATE_ID"),
					Usage:    "The ID for Record you want to update.",
				},
			},
			Aliases:               []string{"set", "update-record"},
			Category:              "dns",
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					params := &cloudflare.UpdateDNSRecordParams{ID: cmd.String("record-id")}
					cfcmd := NewCloudflareCommand(cmd.String("env-file"), cmd.String("domain-name"))
					if cfcmd.Error != nil {
						logger.Error(cfcmd.Error.Error())
						return cfcmd.Error
					}
					if cmd.IsSet("new-content") {
						logger.Debug(cmd.String("new-content"))
						params.Content = cmd.String("new-content")
					}
					if cmd.IsSet("record-name") {
						params.Name = cmd.String("record-name")
					}
					if cmd.IsSet("type") {
						params.Type = cmd.String("type")
					}
					if cmd.IsSet("priority") {
						priority64 := cmd.Uint("priority")
						pr16 := uint16(priority64)
						params.Priority = &pr16
					}
					if cmd.IsSet("ttl") {
						params.TTL = int(cmd.Int("ttl"))
					}
					if cmd.IsSet("proxied") {
						proxied := cmd.Bool("proxied")
						params.Proxied = &proxied
					}
					if cmd.IsSet("comment") {
						comment := cmd.String("comment")
						params.Comment = &comment
					}
					if cmd.IsSet("tags") {
						params.Tags = cmd.StringSlice("tags")
					}
					record := cfcmd.UpdateCloudflareDnsRecord(*params)
					printDnsRecord(record)
					err = cfcmd.Error
				}
				return err
			},
		},
		{
			Name:    "delete",
			Version: versionNumber,
			Authors: cfDnsComandAuthors(),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "rm-record-id",
					Aliases:  []string{"remove-id", "remove-record-id", "delete-record-id"},
					Required: true,
					Sources:  cli.EnvVars("CF_RECORD_DELETE_ID"),
					Usage:    "The ID for Record you want to update.",
				},
			},
			Aliases:               []string{"rm", "remove-record"},
			Category:              "dns",
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				cfcmd := NewCloudflareCommand(cmd.String("env-file"), cmd.String("domain-name"))
				if cfcmd.Error != nil {
					logger.Error(cfcmd.Error.Error())
					return cfcmd.Error
				}
				cfcmd.DeleteCloudflareRecord(cmd.String("rm-record-id"))
				return cfcmd.Error
			},
		},
	}
	return dnsSubCmds
}
