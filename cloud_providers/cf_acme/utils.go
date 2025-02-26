package cf_acme

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

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
