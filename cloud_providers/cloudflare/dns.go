package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type CloudflareDnsZone struct {
	BaseUrl       string         `json:"baseUrl"`
	ZoneId        string         `json:"zoneId"`
	CfToken       string         `json:"cfToken"`
	RecordRequest *DnsRecordReq  `json:"recordRequest"`
	SecretsPath   string         `json:"secretsPath"`
	DnsRecords    []DnsRecordReq `json:"dnsRecords"`
}

type DnsRecordReq struct {
	Content      string    `json:"content"`
	Name         string    `json:"name"`
	Proxied      bool      `json:"proxied"`
	Type         string    `json:"type"`
	Comment      string    `json:"comment"`
	Ttl          int16     `json:"ttl"`
	DnsRecordId  string    `json:"id"`
	ZoneId       string    `json:"zone_id"`
	ZoneName     string    `json:"zone_name"`
	LastModified time.Time `json:"last_modified"`
}

// Define ApiResponse to map the entire JSON response structure
type DnsApiResponse struct {
	Result []DnsRecordReq `json:"result"`
}

func buildRequestUrl(czone *CloudflareDnsZone) string {
	var bUrl bytes.Buffer

	bUrl.WriteString(czone.BaseUrl)
	bUrl.WriteString(czone.ZoneId)
	bUrl.WriteString("/dns_records")
	if czone.RecordRequest.DnsRecordId != "" {
		bUrl.WriteString("/" + czone.RecordRequest.DnsRecordId)
	}

	url := bUrl.String()

	return url
}

func createPayload(record *DnsRecordReq) (io.Reader, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(string(data)), nil
}

func GetCurrentRecords(czone *CloudflareDnsZone) ([]DnsRecordReq, error) {
	//url := "https://api.cloudflare.com/client/v4/zones/[dns_zone_id]/dns_records"

	url := buildRequestUrl(czone)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+czone.CfToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	slog.Debug(fmt.Sprint(res))
	slog.Debug(fmt.Sprint(string(body)))
	slog.Info("Retrieving Current DNS Records")

	// Parse the JSON response
	var apiResponse DnsApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	return apiResponse.Result, nil

}

func GetDnsRecordDetails(czone *CloudflareDnsZone) (DnsRecordReq, error) {
	//url := "https://api.cloudflare.com/client/v4/zones/zone_id/dns_records/dns_record_id"

	url := buildRequestUrl(czone)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+czone.CfToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return DnsRecordReq{}, err
	}

	var dnsDetails DnsRecordReq
	json.Unmarshal(body, &dnsDetails)

	return dnsDetails, nil
}

func CreateDnsRecord(czone *CloudflareDnsZone) (DnsRecordReq, error) {
	url := buildRequestUrl(czone)

	payload, err := createPayload(czone.RecordRequest)
	if err != nil {
		fmt.Println("Error creating payload:", err)
	}
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+czone.CfToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return DnsRecordReq{}, err
	}

	var createRecordRes DnsRecordReq
	json.Unmarshal(body, &createRecordRes)

	return createRecordRes, nil
}

func UpdateDnsRecord(czone *CloudflareDnsZone) (DnsRecordReq, error) {
	url := buildRequestUrl(czone)

	payload, err := createPayload(czone.RecordRequest)
	if err != nil {
		fmt.Println("Error creating payload:", err)
	}
	req, _ := http.NewRequest("PATCH", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+czone.CfToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return DnsRecordReq{}, err
	}

	var updateRecordRes DnsRecordReq
	json.Unmarshal(body, &updateRecordRes)

	return updateRecordRes, nil
}
