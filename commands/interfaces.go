package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudflare/cloudflare-go"
)

type DnsRequestHandler interface {
	cloudflare.UpdateDNSRecordParams | cloudflare.CreateDNSRecordParams
}

func CreateOrUpdateCloudflareDnsRecord[T DnsRequestHandler](envfile string, zoneId string, params T) (cloudflare.DNSRecord, error) {
	record := cloudflare.DNSRecord{}
	api, err := NewCloudflareAPIClient(envfile)
	if err != nil {
		return record, err
	}

	switch v := any(params).(type) {
	case cloudflare.UpdateDNSRecordParams:
		record, err = api.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), v)
	case cloudflare.CreateDNSRecordParams:
		record, err = api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneId), v)
		if err != nil {
			slog.Error("Error updating dns record", slog.String("error", err.Error()), slog.String("RecordID", record.ID))
			return record, err
		}
	default:
		err = fmt.Errorf("unsupported DNS record operation: %T", params)
	}

	return record, err
}
