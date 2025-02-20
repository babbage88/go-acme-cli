package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
)

var logger = pretty.NewCustomLogger(os.Stdout, "DEBUG", 1, "|", true)

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

func GetCloudflareDnsRecordDetails(envfile string, zoneId string, recordId string) (cloudflare.DNSRecord, error) {
	record := cloudflare.DNSRecord{}
	api, err := NewCloudflareAPIClient(envfile)
	if err != nil {
		return record, err
	}

	record, err = api.GetDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), recordId)
	if err != nil {
		slog.Error("Error retrieving record details.", slog.String("error", err.Error()), slog.String("RecordID", record.ID))
		return record, err
	}
	return record, err
}

func GetCloudFlareZoneIdByDomainName(envfile string, domainName string) (string, error) {
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

func UpdateCloudflareDnsRecord(envfile string, zoneId string, recordUpdateParams cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
	record := cloudflare.DNSRecord{}
	api, err := NewCloudflareAPIClient(envfile)
	if err != nil {
		return record, err
	}

	record, err = api.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), recordUpdateParams)
	if err != nil {
		slog.Error("Error updating dns record", slog.String("error", err.Error()), slog.String("RecordID", record.ID))
		return record, err
	}
	return record, err
}

func DeleteCloudFlareDnsRecord(envfile string, zoneId string, recordId string) error {
	api, err := NewCloudflareAPIClient(envfile)
	if err != nil {
		return err
	}
	slog.Info("Deleting Cloudflare DNS Record", slog.String("ZoneID", zoneId), slog.String("RecordID", recordId))
	err = api.DeleteDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), recordId)
	return err
}

func ListCloudflareDnsWithQueryParams(envfile string, domainName string, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error) {
	records := make([]cloudflare.DNSRecord, 0)
	// var recordss = []cloudflare.DNSRecord{}
	result := &cloudflare.ResultInfo{}
	api, err := NewCloudflareAPIClient(envfile)
	if err != nil {
		return records, result, err
	}

	zoneID, err := api.ZoneIDByName(domainName)
	if err != nil {
		slog.Debug("Error retrieving ZoneId for domain name", slog.String("DomainName", domainName))
		return records, result, err
	}

	records, result, err = api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), params)
	if err != nil {
		return records, result, err
	}
	return records, result, nil
}

func GetAllCloudflareDnsRecordByDomain(envfile string, domainName string) ([]cloudflare.DNSRecord, error) {
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

func CreateOrUpdateAcmeDnsRecord(api cloudflare.API, zoneId string, txtRecordName string, txtChalContent string) (cloudflare.DNSRecord, error) {
	var err error
	qryParams := cloudflare.ListDNSRecordsParams{Name: txtRecordName, Type: "TXT"}
	result := &cloudflare.ResultInfo{}
	records := make([]cloudflare.DNSRecord, 0)
	records, result, err = api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneId), qryParams)
	if err != nil {
		slog.Error("error querying DNS records", slog.String("error", err.Error()))
		return cloudflare.DNSRecord{}, err
	}
	if result.Count > 0 {
		for _, v := range records {
			params := cloudflare.UpdateDNSRecordParams{ID: v.ID, Name: txtRecordName, Content: txtChalContent}
			record, retErr := api.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), params)
			if retErr != nil {
				slog.Error("Error updating records", slog.String("error", retErr.Error()))
				return record, retErr
			}
			return record, nil
		}
	}
	params := cloudflare.CreateDNSRecordParams{Name: txtRecordName, Content: txtChalContent}
	record, retErr := api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), params)
	if retErr != nil {
		slog.Error("Error updating records", slog.String("error", retErr.Error()))
		return record, retErr
	}
	return record, nil
}
