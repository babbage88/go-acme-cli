package cloudflare

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log/slog"

	"golang.org/x/crypto/acme"
)

func AcmeRenew() {

	// All the usual account registration prelude
	accountKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: "https://acme-v01.api.letsencrypt.org/directory",
	}

	if _, err := client.Register(context.Background(), &acme.Account{},
		func(tos string) bool {
			slog.Info("Agreeing to ToS", slog.String("tos", tos))
			return true
		}); err != nil {
		slog.Error("Can't register an ACME account: ", slog.String("error", err.Error()))
	}

	// Authorize a DNS name
	authz, err := client.Authorize(context.Background(), "example.org")
	if err != nil {
		slog.Error("can not authorize", slog.String("error", err.Error()))
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
	slog.Info("Creating record.", slog.String("txtLabel", txtLabel), slog.String("txtValue", txtValue))

	// Then the usual: accept the challenge, wait for the authorization ...
	if _, err := client.Accept(context.Background(), chal); err != nil {
		slog.Error("Can't accept challenge", slog.String("error", err.Error()))
	}

	if _, err := client.WaitAuthorization(context.Background(), authz.URI); err != nil {
		slog.Error("Failed authorization.", slog.String("error", err.Error()))
	}

	// Submit certificate request if it suceeded ...
}
