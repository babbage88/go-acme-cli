package cf_acme

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

type ICertRenewalService interface {
	Renew(token string, recurseiveNameservers []string, timeout time.Duration) (CertificateData, error)
}

type InfraCfCustomDNSProvider struct {
	DnsToken             string   `json:"dnsToken"`
	ZoneToken            string   `json:"zoneToken"`
	RecursiveNameServers []string `json:"recursiveNameServers"`
	NewTxtRecordId       string   `json:"newTxtRecordId"`
	CreatedChallengeTXT  bool     `json:"txtCreated"`
	TTL                  int      `json:"ttl"`
	//PropagationTimeout   time.Duration `json:"propagationTimout"`
}

func NewInfraCfCustomDNSProvider(dnsApiToken string, zoneApiToken string, recursiveNameServers []string) (*InfraCfCustomDNSProvider, error) {
	ttl := int(120)
	return &InfraCfCustomDNSProvider{DnsToken: dnsApiToken, ZoneToken: zoneApiToken, RecursiveNameServers: recursiveNameServers, TTL: ttl}, nil
}

func (d *InfraCfCustomDNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	rootDomain := getRootDomain(info.FQDN)
	zoneId, err := GetCloudflareZoneIdFromDomainName(token, rootDomain)
	if err != nil {
		slog.Error("error in InfraCfCustomProvider retrieving zone id", slog.String("error", err.Error()), slog.String("rootDomain", rootDomain))
		return err
	}
	qry, err := CheckRecordNameExists(token, info.FQDN)
	if err != nil {
		slog.Error("error in InfraCfCustomProvider checking of txt record name exists", slog.String("error", err.Error()), slog.String("infoFQDN", info.FQDN))
		return err
	}

	switch qry.Exists {
	case true:
		return fmt.Errorf("txt record already exists, please remove it")
	default:
		params := cloudflare.CreateDNSRecordParams{Name: info.FQDN, Content: info.Value, TTL: d.TTL}
		record, err := CreateCloudflareDnsRecord(token, zoneId, params)
		if err != nil {
			slog.Error("error creating dns record", slog.String("error", err.Error()), slog.String("infofqdn", info.FQDN))
			return err
		}
		d.NewTxtRecordId = record.ID
		d.CreatedChallengeTXT = true

	}
	return err
}

func (d *InfraCfCustomDNSProvider) CleanUp(domain, token, keyAuth string) error {
	zoneId, err := GetCloudflareZoneIdFromDomainName(token, getRootDomain(domain))
	if err != nil {
		slog.Error("error retrieving zone id during cleanup", slog.String("error", err.Error()))
		return err
	}
	if d.CreatedChallengeTXT {
		if len(d.NewTxtRecordId) < 1 {
			return fmt.Errorf("no txt record id specified for cleanup")
		}
		err := DeleteCloudFlareDnsRecord(token, zoneId, d.NewTxtRecordId)
		if err != nil {
			slog.Error("error deleting txt record id during cleanup", slog.String("error", err.Error()), slog.String("id", d.NewTxtRecordId), slog.String("zoneId", zoneId))
			return err
		}
		d.CreatedChallengeTXT = false
		d.NewTxtRecordId = ""
		return err
	}
	return fmt.Errorf("error during cleanup, CreatedChallengeTXT set to false")
}
