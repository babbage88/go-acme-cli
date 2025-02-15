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

func GetCloudflareDnsListByDomainName(envfile string, domainName string) ([]cloudflare.DNSRecord, error) {
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
