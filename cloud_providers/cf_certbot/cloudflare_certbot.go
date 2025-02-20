package cf_certbot

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/acme"
)

func AcmeRenew(envfile string, domainName string, acmeUrl string) error {
	// All the usual account registration prelude
	accountKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: acmeUrl,
		// DirectoryURL: "https://acme-v01.api.letsencrypt.org/directory",
	}

	if _, err := client.Register(context.Background(), &acme.Account{},
		func(tos string) bool {
			slog.Info("Agreeing to ToS", slog.String("tos", tos))
			return true
		}); err != nil {
		slog.Error("Can't register an ACME account: ", slog.String("error", err.Error()))
		return err
	}

	// Authorize a DNS name
	authz, err := client.Authorize(context.Background(), domainName)
	if err != nil {
		slog.Error("can not authorize", slog.String("error", err.Error()))
		return err
	}

	// Find the DNS challenge for this authorization
	var chal *acme.Challenge
	for _, c := range authz.Challenges {
		if c.Type == "dns-01" {
			chal = c
			break
		}
	}
	if chal == nil {
		slog.Error("No DNS challenge was present")
	}

	// Determine the TXT record values for the DNS challenge
	txtLabel := "_acme-challenge." + authz.Identifier.Value
	txtValue, _ := client.DNS01ChallengeRecord(chal.Token)

	// Initialize cloudflare api and create or update txt record.
	slog.Info("Creating record.", slog.String("txtLabel", txtLabel), slog.String("txtValue", txtValue))
	_, err = handleAcmeDnsChallenge(envfile, domainName, txtLabel, txtValue)
	if err != nil {
		slog.Error("Error handling DNS challenge", slog.String("error", err.Error()))
		return err
	}

	// Then the usual: accept the challenge, wait for the authorization ...
	if _, err := client.Accept(context.Background(), chal); err != nil {
		slog.Error("Can't accept challenge", slog.String("error", err.Error()))
		return err
	}

	if _, err := client.WaitAuthorization(context.Background(), authz.URI); err != nil {
		slog.Error("Failed authorization.", slog.String("error", err.Error()))
		return err
	}

	// Submit certificate request if it suceeded ...
	// Generate a new private key for the certificate
	certKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		slog.Error("failed to generate cert key.", slog.String("error", err.Error()))
		return err
	}
	// Request the certificate
	csr, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		DNSNames: []string{domainName},
	}, certKey)
	if err != nil {
		slog.Error("failed to create CSR.", slog.String("error", err.Error()))
		return err
	}

	// Encode CSR to PEM
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr})

	// Finalize order and get certificate
	certDER, _, err := client.CreateOrderCert(context.Background(), authz.URI, csrPEM, true)
	if err != nil {
		slog.Error("failed to get certificate.", slog.String("error", err.Error()))
		return err
	}

	// Save certificate
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER[0]})
	fmt.Println("Certificate:\n", string(certPEM))
	return nil
}

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
