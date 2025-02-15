package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

const versionNumber = "v1.0.0"

type GoInfraCli struct {
	Name        string       `json:"name"`
	BaseCommand *cli.Command `json:"baseCmd"`
	Version     string       `json:"version"`
}

type CloudflareCommandUtils struct {
	ZomeId    string          `json:"zoneId"`
	EnvFile   string          `json:"envFile"`
	Error     error           `json:"error"`
	ApiClient *cloudflare.API `json:"clouflareApi"`
}

func NewCloudflareCommand(envfile string, domainName string) *CloudflareCommandUtils {
	cfcmd := &CloudflareCommandUtils{EnvFile: envfile}
	cfcmd.Error = godotenv.Load(cfcmd.EnvFile)
	cfcmd.NewApiClient()
	if cfcmd.Error == nil {
		cfcmd.ZomeId, cfcmd.Error = cfcmd.ApiClient.ZoneIDByName(domainName)
	}

	return cfcmd
}

func (cf *CloudflareCommandUtils) NewApiClient() {

	cf.Error = godotenv.Load(cf.EnvFile)
	cf.ApiClient, cf.Error = cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_DNS_API_TOKEN"))
}

type UrFaveCliDocumentationSucks struct {
	Name  string `json:"authorName"`
	Email string `json:"email"`
}

func (cfcmd CloudflareCommandUtils) UpdateCloudflareDnsRecord(recordUpdateParams cloudflare.UpdateDNSRecordParams) cloudflare.DNSRecord {
	record := cloudflare.DNSRecord{}

	paramb, err := json.Marshal(recordUpdateParams)
	if err != nil {
		pretty.PrintErrorf("error marshaling json %s", err.Error())
	}
	pretty.Print(string(paramb))
	record, cfcmd.Error = cfcmd.ApiClient.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(cfcmd.ZomeId), recordUpdateParams)
	return record
}

func (cfcmd *CloudflareCommandUtils) GetDnsRecord(recordId string) cloudflare.DNSRecord {
	record := cloudflare.DNSRecord{}
	record, cfcmd.Error = cfcmd.ApiClient.GetDNSRecord(context.Background(), cloudflare.ZoneIdentifier(cfcmd.ZomeId), recordId)
	return record

}

func (cfcmd *CloudflareCommandUtils) CreateOrUpdateDNSRecord(params any) cloudflare.DNSRecord {
	record := cloudflare.DNSRecord{}

	switch v := any(params).(type) {
	case cloudflare.UpdateDNSRecordParams:
		record, cfcmd.Error = createOrUpdateCloudflareDnsRecord(*cfcmd.ApiClient, cfcmd.ZomeId, v)
	case cloudflare.CreateDNSRecordParams:
		record, cfcmd.Error = createOrUpdateCloudflareDnsRecord(*cfcmd.ApiClient, cfcmd.ZomeId, v)
	default:
		cfcmd.Error = fmt.Errorf("unsupported DNS record operation: %T", params)
	}

	return record
}

func (cfcmd CloudflareCommandUtils) DeleteCloudflareRecord(recordId string) {
	cfcmd.Error = cfcmd.ApiClient.DeleteDNSRecord(context.Background(), cloudflare.ZoneIdentifier(cfcmd.ZomeId), recordId)
	if cfcmd.Error == nil {
		msg := fmt.Sprintf("DNS RecordID: %s in Zone: %s has been deleted succesfully", recordId, cfcmd.ZomeId)
		logger.Info(msg)
	}
}

func (author *UrFaveCliDocumentationSucks) String() string {
	return fmt.Sprintf("Name: %s Email: %s", author.Name, author.Email)
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

func createOrUpdateCloudflareDnsRecord[T DnsRequestHandler](api cloudflare.API, zoneId string, params T) (cloudflare.DNSRecord, error) {
	record := cloudflare.DNSRecord{}
	var err error = nil
	switch v := any(params).(type) {
	case cloudflare.UpdateDNSRecordParams:
		record, err = api.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), v)
		if err != nil {
			msg := fmt.Sprintf("Error updating DNS record %s in Zone: %s err: %s", v.ID, zoneId, err.Error())
			logger.Error(msg)
			return record, err
		}
		return record, err
	case cloudflare.CreateDNSRecordParams:
		record, err = api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), v)
		if err != nil {
			msg := fmt.Sprintf("Error updating DNS record %s in Zone: %s err: %s", v.Name, zoneId, err.Error())
			logger.Error(msg)
			return record, err
		}
	default:
		err = fmt.Errorf("unsupported DNS record operation %T. Must use cloudflare.UpdateDNSRecordParams or CreateDNSRecordParams.", params)
	}
	return record, err
}
