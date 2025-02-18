package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/babbage88/go-acme-cli/database/infracli_db"
	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

const versionNumber = "v1.0.0"

type CloudflareCommandUtils struct {
	ZomeId    string          `json:"zoneId"`
	ZoneName  string          `json:"zoneName"`
	EnvFile   string          `json:"envFile"`
	Error     error           `json:"error"`
	ApiClient *cloudflare.API `json:"clouflareApi"`
	DbConn    *sql.DB         `json:"db"`
}

func NewCloudflareCommand(envfile string, domainName string) *CloudflareCommandUtils {
	cfcmd := &CloudflareCommandUtils{EnvFile: envfile, ZoneName: domainName}
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

func (cfcmd *CloudflareCommandUtils) ListDNSRecords(params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo) {
	records := []cloudflare.DNSRecord{}
	results := &cloudflare.ResultInfo{}
	records, results, cfcmd.Error = cfcmd.ApiClient.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(cfcmd.ZomeId), params)
	return records, results
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

func (cfcmd *CloudflareCommandUtils) PrintDnsRecordsTable(records []cloudflare.DNSRecord) {
	var colorInt int32 = 97
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(tw, "\x1b[1;%dm%s\t%s\t%s\t%s\t%s\t%s\t%s\x1b[0m\n", colorInt, "ID", "Name", "Content", "Type", "CreatedOn", "ModifiedOn", "Comment")
	fmt.Fprintf(tw, "\x1b[1;%dm--\t----\t-------\t----\t---------\t----------\t-------\x1b[0m\n", colorInt)
	for _, v := range records {
		switch v.Type {
		case "A":
			colorInt = int32(96)
		case "CNAME":
			colorInt = int32(92)
		default:
			colorInt = int32(97)
		}
		fmt.Fprintf(tw, "\x1b[1;%dm%s\t%s\t%s\t%s\t%s\t%s\t%s\x1b[0m\n", colorInt, v.ID, v.Name, v.Content, v.Type, pretty.DateTimeSting(v.CreatedOn), pretty.DateTimeSting(v.ModifiedOn), v.Comment)
	}
	tw.Flush()
	fmt.Printf("\x1b[1;%dm\nFound %d records in ZoneID: %s Name: %s\x1b[0m\n", colorInt, len(records), cfcmd.ZomeId, cfcmd.ZoneName)
}

func (cfcmd *CloudflareCommandUtils) DeleteCloudflareRecord(recordId string) {
	cfcmd.Error = cfcmd.ApiClient.DeleteDNSRecord(context.Background(), cloudflare.ZoneIdentifier(cfcmd.ZomeId), recordId)
	if cfcmd.Error == nil {
		msg := fmt.Sprintf("DNS RecordID: %s in Zone: %s has been deleted succesfully", recordId, cfcmd.ZomeId)
		logger.Info(msg)
	}
}

func (cfcmd *CloudflareCommandUtils) PrintCommandResultAsJson(result any) string {
	fmt.Println()
	response, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		cfcmd.Error = err
		msg := fmt.Sprintf("error marshaling result into json. error: %s", err.Error())
		logger.Error(msg)
	}
	recordsJson := fmt.Sprintf("%s\n", (string(response)))
	pretty.Print(recordsJson)

	return recordsJson
}

func (cfcmd *CloudflareCommandUtils) InitializeDatabaseConnection() {
	cfcmd.Error = godotenv.Load(cfcmd.EnvFile)

	if cfcmd.Error != nil {
		msg := fmt.Sprintf("error loading .env: %s", cfcmd.Error.Error())
		logger.Error(msg)
	}

	dbfile := os.Getenv("SQLITE_DB_PATH")

	if dbfile == "" {
		logger.Error("SQLITE_DB_PATH is not set")
	}

	cfcmd.DbConn, cfcmd.Error = sql.Open("sqlite3", dbfile)
	if cfcmd.Error != nil {
		log.Fatalf("Failed to open database: %v", cfcmd.Error.Error())
	}
}

func (cfcmd *CloudflareCommandUtils) GetZonesFromDb() []infracli_db.DnsZone {
	if cfcmd.DbConn == nil {
		cfcmd.InitializeDatabaseConnection()
	}
	queries := infracli_db.New(cfcmd.DbConn)
	zones, err := queries.GetZonesFromDb(context.Background())
	cfcmd.Error = err
	return zones
}

func (cfcmd *CloudflareCommandUtils) CreateZoneInDb() {
	if cfcmd.DbConn == nil {
		cfcmd.InitializeDatabaseConnection()
	}
	params := infracli_db.CreateDnsZoneParams{ZoneUid: cfcmd.ZomeId, DomainName: cfcmd.ZoneName}
	queries := infracli_db.New(cfcmd.DbConn)

	cfcmd.Error = queries.CreateDnsZone(context.Background(), params)
	if cfcmd.Error != nil {
		log.Fatalf("Failed to create DNS zone: %v", cfcmd.Error.Error())
	}
}

func (cfcmd *CloudflareCommandUtils) PrintZoneIdTable() error {
	var colorInt int32 = 92
	tw := tabwriter.NewWriter(os.Stdout, 5, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintf(tw, "\x1b[1;%dm", colorInt)
	fmt.Fprintf(tw, "\tZoneName\t\tZoneID\n")
	fmt.Fprintf(tw, "--------\t\t------\n")
	fmt.Fprintf(tw, "%s\t\t%s\n", cfcmd.ZoneName, cfcmd.ZomeId)
	fmt.Fprintf(tw, "\x1b[0m")
	err := tw.Flush()
	return err
}

func (cfcmd *CloudflareCommandUtils) PrintDnsZoneDbRecords(zones []infracli_db.DnsZone) error {
	// var colorInt int32 = 97
	tw := tabwriter.NewWriter(os.Stdout, 5, 0, 1, ' ', tabwriter.AlignRight)
	// fmt.Fprintf(tw, "\x1b[1;%dm", colorInt)
	fmt.Fprintf(tw, "ID\t\tZoneName\t\tZoneID\n")
	fmt.Fprintf(tw, "--\t\t--------\t\t------\n")

	for _, v := range zones {
		fmt.Printf("%d %s %s\n", v.ID, v.DomainName, v.ZoneUid)
		fmt.Fprintf(tw, "%d\t\t%s\t\t%s\n", v.ID, v.DomainName, v.ZoneUid)
	}
	// fmt.Fprintf(tw, "\x1b[0m")
	err := tw.Flush()
	return err
}

func (author *UrFaveCliDocumentationSucks) String() string {
	return fmt.Sprintf("Name: %s Email: %s", author.Name, author.Email)
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
		err = fmt.Errorf("unsupported DNS record operation %T. Must use cloudflare.UpdateDNSRecordParams or CreateDNSRecordParams", params)
	}
	return record, err
}
