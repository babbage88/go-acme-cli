package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/cloudflare/cloudflare-go"
)

func printDnsRecord(record cloudflare.DNSRecord) {
	var colorInt int32 = 97

	switch record.Type {
	case "A":
		colorInt = int32(96)
	case "CNAME":
		colorInt = int32(92)
	default:
		colorInt = int32(97)
	}
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(tw, "\x1b[1;%dm%s\t%s\t%s\t%s\t%s\t%s\t%s\x1b[0m\n", colorInt, "ID", "Name", "Content", "Type", "CreatedOn", "ModifiedOn", "Comment")
	fmt.Fprintf(tw, "\x1b[1;%dm--\t----\t-------\t----\t---------\t----------\t-------\x1b[0m\n", colorInt)
	fmt.Fprintf(tw, "\x1b[1;%dm%s\t%s\t%s\t%s\t%s\t%s\t%s\x1b[0m\n", colorInt, record.ID, record.Name, record.Content, record.Type, pretty.DateTimeSting(record.CreatedOn), pretty.DateTimeSting(record.ModifiedOn), record.Comment)
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
