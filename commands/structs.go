package commands

import (
	"context"
	"fmt"
	"os"

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

	record, cfcmd.Error = cfcmd.ApiClient.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(cfcmd.ZomeId), recordUpdateParams)
	return record
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
