package cf_certbot

import (
	"context"
	"log/slog"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
)

func initializeCloudflareAPIClient(envfile string) (*cloudflare.API, error) {
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

func createOrUpdateAcmeDnsRecord(api cloudflare.API, zoneId string, txtRecordName string, txtChalContent string) (cloudflare.DNSRecord, error) {
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

func handleAcmeDnsChallenge(envfile string, domainName string, txtLabel string, txtValue string) (cloudflare.DNSRecord, error) {
	record := cloudflare.DNSRecord{}
	api, err := initializeCloudflareAPIClient(envfile)
	if err != nil {
		slog.Error("error opening cloudflare api client", slog.String("error", err.Error()))
		return record, err
	}
	zoneID, err := api.ZoneIDByName(domainName)
	if err != nil {
		slog.Error("error retrieving cloudflare zone id from api client", slog.String("error", err.Error()))
		return record, err
	}

	slog.Info("creating or updatin dns record", slog.String("zoneId", zoneID), slog.String("Name", txtLabel))
	record, err = createOrUpdateAcmeDnsRecord(*api, zoneID, txtLabel, txtValue)
	if err != nil {
		slog.Error("error creating or updating txt record during cert renew request", slog.String("error", err.Error()), slog.String("txtLabel", txtLabel))
		return record, err
	}

	return record, nil
}
