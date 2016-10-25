package ca


import (
	"google.golang.org/grpc"
	pb "github.com/ethereum/go-ethereum/crypto/caserver/ca/protos"
	"time"
	"crypto/x509"
	"crypto/ecdsa"
	"encoding/pem"
	"io/ioutil"
	"strings"
	"os"
	"fmt"
	"golang.org/x/net/context"
	"github.com/spf13/viper"
)

func GetClientConn() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	address := viper.GetString("caserver.address") + viper.GetString("caserver.port")

	return grpc.Dial(address, opts...)
}

func GetWhitelistClient() (*grpc.ClientConn, pb.WhitelistClient, error) {
	conn, err := GetClientConn()
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewWhitelistClient(conn)
	return conn, client, nil
}

func GetCAClient() (*grpc.ClientConn, pb.CAClient, error) {
	conn, err := GetClientConn()
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewCAClient(conn)
	return conn, client, nil
}

func GetCACertificate() ([]byte, error) {

	sock, caClient, err := GetCAClient()
	if err != nil {
		return nil, err
	}
	defer sock.Close()

	resp, err := caClient.GetCACertificate(context.Background(), &pb.NoParam{})
	if err != nil {
		return nil, fmt.Errorf("could not GetCACertificate: %v", err)
	}

	return resp.In, nil
}

func IssueCertificate(pubKey *ecdsa.PublicKey, name, path string) (*x509.Certificate, error) {
	name = strings.Replace(name, "/", "_", -1)
	host, _ := os.Hostname()
	name += host

	sock, caClient, err := GetCAClient()
	if err != nil {
		return nil, err
	}
	defer sock.Close()

	raw, _ := x509.MarshalPKIXPublicKey(pubKey)
	cooked := pem.EncodeToMemory(
		&pem.Block{
			Type:  "ECDSA PUBLIC KEY",
			Bytes: raw,
		})

	req := &pb.CertificateRequest{
		In:	[]byte(cooked),
		Name:	name}

	resp, err := caClient.IssueCertificate(context.Background(), req)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path+"/cakeystore/"+name+".cert", resp.In, 0644)
	if err != nil {
		caLogger.Panic(err)
	}

	block, _ := pem.Decode(resp.In)
	if block == nil {
		return nil, fmt.Errorf("certificate data error.")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	return cert, err
}

func ReadCACertificate(name, path string) (*x509.Certificate, error) {
	caLogger.Debug("Reading CA certificate.")

	name = strings.Replace(name, "/", "_", -1)
	host, _ := os.Hostname()
	name += host

	path = path + "/cakeystore/" + name + ".cert"

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	cooked, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(cooked)
	if block == nil {
		return nil, fmt.Errorf("certificate data error.")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	return cert, err
}