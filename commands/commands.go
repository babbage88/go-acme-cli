package commands

import (
	"context"
	"fmt"

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
			Aliases: []string{"record-content", "c"},
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
			Name:    "list",
			Version: versionNumber,
			Authors: cfDnsComandAuthors(),
			Aliases: []string{"list-records", "ls"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "qry-record-content",
					Aliases: []string{"qry-content"},
					Usage:   "Record content for list query.",
				},
				&cli.StringFlag{
					Name:    "qry-record-type",
					Aliases: []string{"qry-type"},
					Usage:   "Record type for list query.",
				},
				&cli.BoolFlag{
					Name:    "qry-record-proxied",
					Aliases: []string{"qry-proxied"},
					Usage:   "Return proxied records only.",
				},
				&cli.StringFlag{
					Name:    "qry-record-name",
					Aliases: []string{"qry-name"},
					Usage:   "Record name for list query.",
				},
				&cli.StringFlag{
					Name:    "qry-record-comment",
					Aliases: []string{"qry-comment"},
					Usage:   "Record comment for list query.",
				},
				&cli.StringSliceFlag{
					Name:    "qry-record-tags",
					Aliases: []string{"qry-tags"},
					Usage:   "Record tags for list query.",
				},
				&cli.UintFlag{
					Name:    "qry-record-priority",
					Aliases: []string{"query-priority"},
					Usage:   "Return list of all records that match query priority",
				},
				&cli.BoolFlag{
					Name:    "list-qry",
					Aliases: []string{"query-records"},
					Usage:   "Return list of all records that match query params",
					Value:   false,
				},
				&cli.BoolFlag{
					Name:    "print-json",
					Aliases: []string{"show-json"},
					Usage:   "Return list of all records that match query params",
					Value:   false,
				},
			},
			Category:              "dns",
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				cfcmd := NewCloudflareCommand(cmd.String("env-file"), cmd.String("domain-name"))
				var params = &cloudflare.ListDNSRecordsParams{}
				if cfcmd.Error != nil {
					logger.Error(cfcmd.Error.Error())
					return cfcmd.Error
				}
				if cmd.Bool("list-qry") {
					if cmd.IsSet("qry-record-content") {
						params.Content = cmd.String("qry-record-content")
					}
					if cmd.IsSet("qry-record-name") {
						params.Name = cmd.String("qry-record-name")
					}
					if cmd.IsSet("qry-record-type") {
						params.Type = cmd.String("qry-record-type")
					}
					if cmd.IsSet("qry-record-priority") {
						priority64 := cmd.Uint("qry-record-priority")
						pr16 := uint16(priority64)
						params.Priority = &pr16
					}
					if cmd.IsSet("qry-record-proxied") {
						proxied := cmd.Bool("qry-record-proxied")
						params.Proxied = &proxied
					}
					if cmd.IsSet("qry-record-comment") {
						params.Comment = cmd.String("qry-record-comment")
					}
					if cmd.IsSet("qry-record-tags") {
						params.Tags = cmd.StringSlice("tags")
					}
					records, _ := cfcmd.ListDNSRecords(*params)
					cfcmd.PrintDnsRecordsTable(records)
					return cfcmd.Error
				}
				records, _ := cfcmd.ListDNSRecords(*params)
				if cmd.Bool("print-json") {
					cfcmd.PrintCommandResultAsJson(records)
					return cfcmd.Error
				}
				cfcmd.PrintDnsRecordsTable(records)
				return cfcmd.Error
			},
		},
		{
			Name:    "get",
			Version: versionNumber,
			Authors: cfDnsComandAuthors(),
			Aliases: []string{"get-record", "cat"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "get-record-id",
					Aliases: []string{"qry-record-id"},
					Usage:   "The ID for Record you want to get details for.",
				},
			},
			Category:              "dns",
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					cfcmd := NewCloudflareCommand(cmd.String("env-file"), cmd.String("domain-name"))
					if cfcmd.Error != nil {
						logger.Error(cfcmd.Error.Error())
						return cfcmd.Error
					}
					record := cfcmd.GetDnsRecord(cmd.String("get-record-id"))
					printDnsRecordAsJson(record)
					return cfcmd.Error
				}
				cfcmd := NewCloudflareCommand(cmd.String("env-file"), cmd.String("domain-name"))
				if cfcmd.Error != nil {
					logger.Error(cfcmd.Error.Error())
					return cfcmd.Error
				}
				record := cfcmd.GetDnsRecord(cmd.Args().Get(0))
				printDnsRecordAsJson(record)
				return cfcmd.Error
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
					record := cfcmd.CreateOrUpdateDNSRecord(*params)
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
					Name:    "rm-record-id",
					Aliases: []string{"remove-id", "remove-record-id", "delete-record-id"},
					Sources: cli.EnvVars("CF_RECORD_DELETE_ID"),
					Usage:   "The ID for Record you want to update.",
				},
			},
			Aliases:               []string{"rm", "remove-record"},
			Category:              "dns",
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					cfcmd := NewCloudflareCommand(cmd.String("env-file"), cmd.String("domain-name"))
					if cfcmd.Error != nil {
						logger.Error(cfcmd.Error.Error())
						return cfcmd.Error
					}
					cfcmd.DeleteCloudflareRecord(cmd.String("rm-record-id"))
					return cfcmd.Error
				}
				cfcmd := NewCloudflareCommand(cmd.String("env-file"), cmd.String("domain-name"))
				if cfcmd.Error != nil {
					logger.Error(cfcmd.Error.Error())
					return cfcmd.Error
				}
				cfcmd.DeleteCloudflareRecord(cmd.Args().Get(0))
				return cfcmd.Error
			},
		},
		{
			Name:                  "create",
			Version:               versionNumber,
			Authors:               cfDnsComandAuthors(),
			Aliases:               []string{"add", "create-record", "new-record"},
			Category:              "dns",
			EnableShellCompletion: true,
			Action: func(ctx context.Context, cmd *cli.Command) (err error) {
				if cmd.NArg() == 0 {
					params := &cloudflare.CreateDNSRecordParams{}
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
						params.Comment = cmd.String("comment")
					}
					if cmd.IsSet("tags") {
						params.Tags = cmd.StringSlice("tags")
					}
					record := cfcmd.CreateOrUpdateDNSRecord(*params)
					if cfcmd.Error != nil {
						msg := fmt.Sprintf("Error creating new DNS record: %s in Zone: %s error: %s", cmd.String("record-name"), cfcmd.ZomeId, cfcmd.Error.Error())
						logger.Error(msg)
					}
					printDnsRecord(record)
					err = cfcmd.Error
				}
				return err
			},
		},
	}
	return dnsSubCmds
}
