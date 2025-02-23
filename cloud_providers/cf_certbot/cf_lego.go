package cf_certbot

import (
	"archive/zip"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/babbage88/go-acme-cli/internal/pretty"
	"github.com/babbage88/go-acme-cli/storage/goinfra_minio"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	lego_cloudflare "github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
	"github.com/joho/godotenv"
)

type CertificateData struct {
	DomainNames     []string `json:"domainName"`
	CertPEM         string   `json:"cert_pem"`
	ChainPEM        string   `json:"chain_pem"`
	Fullchain       string   `json:"fullchain_pem"`
	FullchainAndKey string   `json:"fullchain_and_key"`
	PrivKey         string   `json:"priv_key"`
	ZipDir          string   `json:"zipDir"`
	S3DownloadUrl   string   `json:"s3DownloadUrl"`
}

type CertificateRenewalRequest struct {
	EnvFile     string   `json:"envFile"`
	DomainNames []string `json:"domainName"`
	AcmeEmail   string   `json:"acmeEmail"`
	AcmeUrl     string   `json:"acmeUrl"`
	SaveZip     bool     `json:"saveZip"`
	ZipDir      string   `json:"zipDir"`
	PushS3      bool     `json:"pushS3"`
}

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (c *CertificateRenewalRequest) RenewCertWithDns() (CertificateData, error) {
	certdata := &CertificateData{DomainNames: c.DomainNames}
	err := godotenv.Load(c.EnvFile)
	if err != nil {
		slog.Error("Error loading .env file", slog.String("error", err.Error()))
		return *certdata, err
	}
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		slog.Error("error creating private key", slog.String("error", err.Error()))
		return *certdata, err
	}

	acmeUser := AcmeUser{
		Email: c.AcmeEmail,
		key:   privateKey,
	}

	config := lego.NewConfig(&acmeUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = c.AcmeUrl
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		slog.Error("error creating client", slog.String("error", err.Error()))
		return *certdata, err
	}

	provider, err := lego_cloudflare.NewDNSProvider()
	recursiveServers := dns01.AddRecursiveNameservers(([]string{"1.1.1.1:53", "1.0.0.1:53"}))
	timeout := dns01.AddDNSTimeout(60 * time.Second)
	//pt := dns01.DisableAuthoritativeNssPropagationRequirement()

	if err != nil {
		if err != nil {
			slog.Error("error initializing cloudflare DNS challenge provider", slog.String("error", err.Error()))
			return *certdata, err
		}
	}

	err = client.Challenge.SetDNS01Provider(provider, recursiveServers, timeout)
	if err != nil {
		slog.Error("Failed to set DNS challenge.", slog.String("error", err.Error()))
		return *certdata, err
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		slog.Error("Error creating registration", slog.String("error", err.Error()))
		return *certdata, err
	}
	acmeUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: c.DomainNames,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}
	cert := string(certificates.Certificate)
	privKey := string(certificates.PrivateKey)
	issuerCA := string(certificates.IssuerCertificate)
	fullChain := fmt.Sprint(cert, issuerCA)
	certdata.CertPEM = cert
	certdata.PrivKey = privKey
	certdata.Fullchain = fullChain
	certdata.FullchainAndKey = fmt.Sprint(fullChain, privKey)

	if c.SaveZip {
		err = saveToZip(c.ZipDir, certificates.Certificate, certificates.PrivateKey, certificates.IssuerCertificate)
		if err != nil {
			slog.Error("error saving zip", slog.String("error", err.Error()))
		}
		certdata.ZipDir = c.ZipDir
	}

	if c.PushS3 {
		objName := fmt.Sprint(strings.TrimLeft(c.DomainNames[0], "*"), "certs.zip")
		s3client, initErr := goinfra_minio.NewS3ClientFromEnv(c.EnvFile)
		if initErr != nil {
			err = initErr
			slog.Error("error in pushS3 stiep", slog.String("error", err.Error()))
		}
		_, pushErr := s3client.PushFileToDefaultBucket(objName, c.ZipDir)
		if pushErr != nil {
			err = pushErr
			slog.Error("error pushing file to s3", slog.String("error", err.Error()), slog.String("sourceFile", c.ZipDir))
		}
		expiry := 15 * time.Minute
		presignedUrl, err := s3client.Client.PresignedGetObject(context.Background(), s3client.DefaultBucketName, objName, expiry, nil)
		if err != nil {
			slog.Error("Error generating presigned download URL", slog.String("error", err.Error()))
		}
		certdata.S3DownloadUrl = presignedUrl.String()
	}
	return *certdata, err
}

func (c *CertificateRenewalRequest) CliRenewal() (CertificateData, error) {
	certData, err := c.RenewCertWithDns()
	if err != nil {
		slog.Error("error renewing certificate", slog.String("error", err.Error()))
		return CertificateData{DomainNames: c.DomainNames}, err
	}
	printJson(certData)

	if c.SaveZip {
		pretty.Printf("Zip File location: %s", c.ZipDir)
	}
	return certData, err
}

func printJson(data any) {
	response, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		slog.Error("error marshaling response.", slog.String("error", err.Error()))
	}
	fmt.Printf("%s\n", string(response))
	fmt.Println()
}

func saveToZip(filename string, certPEM []byte, keyPEM []byte, issuerCA []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	files := map[string][]byte{
		"certificate.pem": certPEM,
		"private_key.pem": keyPEM,
		"issuer_ca.pem":   issuerCA,
	}

	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		_, err = writer.Write(content)
		if err != nil {
			return err
		}
	}

	return nil
}
