package ca

import (
	"testing"
	pb "github.com/ethereum/go-ethereum/crypto/caserver/ca/protos"
	"github.com/ethereum/go-ethereum/Godeps/_workspace/src/golang.org/x/net/context"
)

const (
	pub = `-----BEGIN ECDSA PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEzqR158ptAz23PsGiKeAAQfdgaUP3
1j7hyO4lqc+b1rUwsCW9ED5P94ysslg6e75MT6UCKYLqRYlIr3bOqfT51w==
-----END ECDSA PUBLIC KEY-----`

	client2Cert  = `-----BEGIN CERTIFICATE-----
MIIByjCCAXCgAwIBAgIBATAKBggqhkjOPQQDAzBRMR4wHAYDVQQGExVwa2kuY2Eu
c2hhbmdoYWkuY2hpbmExGjAYBgNVBAoTEXBraS5jYS5ibG9ja2NoYWluMRMwEQYD
VQQDEwpCbG9ja2NoYWluMB4XDTE2MTAxMTA2MDQyOVoXDTE3MDEwOTA2MDQyOVow
SjEeMBwGA1UEBhMVcGtpLmNhLnNoYW5naGFpLmNoaW5hMRowGAYDVQQKExFwa2ku
Y2EuYmxvY2tjaGFpbjEMMAoGA1UEAxMDcHViMFkwEwYHKoZIzj0CAQYIKoZIzj0D
AQcDQgAEzqR158ptAz23PsGiKeAAQfdgaUP31j7hyO4lqc+b1rUwsCW9ED5P94ys
slg6e75MT6UCKYLqRYlIr3bOqfT516NAMD4wDgYDVR0PAQH/BAQDAgKEMAwGA1Ud
EwEB/wQCMAAwDQYDVR0OBAYEBAECAwQwDwYDVR0jBAgwBoAEAQIDBDAKBggqhkjO
PQQDAwNIADBFAiATi8CDwtspePyUWSzjHNvY+7PKK2UWfK5haRFrY1hRUQIhANLv
Y6T+HtBlicOlulo7aX//Za07RM3iWRdKREHktVOR
-----END CERTIFICATE-----`

	rootCert = `-----BEGIN CERTIFICATE-----
MIIBwjCCAWmgAwIBAgIBATAKBggqhkjOPQQDAzBRMR4wHAYDVQQGExVwa2kuY2Eu
c2hhbmdoYWkuY2hpbmExGjAYBgNVBAoTEXBraS5jYS5ibG9ja2NoYWluMRMwEQYD
VQQDEwpCbG9ja2NoYWluMB4XDTE2MTAxMTA0MTAxNloXDTE3MDEwOTA0MTAxNlow
UTEeMBwGA1UEBhMVcGtpLmNhLnNoYW5naGFpLmNoaW5hMRowGAYDVQQKExFwa2ku
Y2EuYmxvY2tjaGFpbjETMBEGA1UEAxMKQmxvY2tjaGFpbjBZMBMGByqGSM49AgEG
CCqGSM49AwEHA0IABDlS2GWIWcmd+gXhjvqBoNFpHKDPjTSi4vqoRIBSa5brUpzR
T3jb7trVUDfuTQL4EqUFJfnZ1xxj2g/G9AqtiFWjMjAwMA4GA1UdDwEB/wQEAwIC
hDAPBgNVHRMBAf8EBTADAQH/MA0GA1UdDgQGBAQBAgMEMAoGCCqGSM49BAMDA0cA
MEQCIE+odaiwtt5Ug9xQKimu6yJhIpNUzsBE92kghF8eOim3AiB0cfMn+UutRv7x
mzCHolJq0sWyYrscotfVMpvKw5zaNw==
-----END CERTIFICATE-----`
)

func TestIssueCertificate(t *testing.T) {
	sock, caClient, err := GetCAClient()
	if err != nil {
		t.Fatalf("Error executing test: %v", err)
	}
	defer sock.Close()

	req := &pb.CertificateRequest{
		In:	[]byte(pub),
		Name:	"pub"}

	resp, err := caClient.IssueCertificate(context.Background(), req)

	if err != nil {
		t.Fatalf("Error executing test: %v", err)
	}
	t.Logf("IssueCertificate: %s", resp.In)
}

func TestGetCACertificate(t *testing.T) {
	sock, caClient, err := GetCAClient()
	if err != nil {
		t.Fatalf("Error executing test: %v", err)
	}
	defer sock.Close()

	resp, err := caClient.GetCACertificate(context.Background(), &pb.NoParam{})

	if err != nil {
		t.Fatalf("Error executing test: %v", err)
	}
	t.Logf("GetCACertificate: %s", resp.In)
}

func TestVerifySignature(t *testing.T) {
	sock, caClient, err := GetCAClient()
	if err != nil {
		t.Fatalf("Error executing test: %v", err)
	}
	defer sock.Close()

	req := &pb.CertificateData{
		Cert:	[]byte(client2Cert),
		Root:	[]byte(rootCert)}

	resp, err := caClient.VerifySignature(context.Background(), req)

	if err != nil {
		t.Fatalf("Error executing test: %v", err)
	}
	t.Logf("VerifySignature: %s", resp.Valid)
}