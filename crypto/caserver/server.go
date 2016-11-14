package main


import (
	"net"

	"github.com/op/go-logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "github.com/ethereum/go-ethereum/crypto/caserver/ca/protos"
	"log"
	"github.com/ethereum/go-ethereum/crypto/caserver/ca"
	"os"
	"strings"
	"github.com/spf13/viper"
	"fmt"
)

const (
	envPrefix = "BC_CONF"
)

var slogger = logging.MustGetLogger("server")
var cap *ca.CA


type whitelistServer struct{}

func (s *whitelistServer) GetWhitelist(ctx context.Context, in *pb.NoParam) (*pb.IPList, error) {
	res := &pb.IPList{}
	res.Ip = make([]string, 2)
	res.Ip[0] = "127.0.0.1"
	res.Ip[1] = "192.168.0.1"

	return res, nil
}

type CAServer struct{}

func (s *CAServer)IssueCertificate(ctx context.Context, cr *pb.CertificateRequest) (*pb.CertificateReply, error) {
	if cap == nil {
		return nil, nil
	}

	reply := pb.CertificateReply{}

	name := strings.Replace(cr.Name, "/", "_", -1)

	if cert, err := cap.IssueCertificate(cr.In, name, ca.Validator); err != nil {
		slogger.Panicf("Failed IssueCertificate [%s]", err)
		return nil, err
	}else {
		reply.In = cert
	}

	return &reply, nil
}

func (s *CAServer)GetCACertificate(ctx context.Context, np *pb.NoParam) (*pb.CertificateReply, error) {
	if cap == nil {
		return nil, nil
	}

	reply := pb.CertificateReply{}
	reply.In = cap.GetCACertificate()

	return &reply, nil
}

func (s *CAServer)VerifySignature(ctx context.Context, certData *pb.CertificateData) (*pb.SignatureValid, error) {
	valid := pb.SignatureValid{}

	err := ca.VerifySignature(certData.Cert, certData.Root)
	valid.Valid = err == nil

	return &valid, err
}

func main() {

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetConfigName("properties")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./common/")
	viper.AddConfigPath("../../common/")

	err := viper.ReadInConfig()
	if err != nil {
		slogger.Panicf("Fatal error when reading config file: %s", err)
	}

	fmt.Println(viper.GetString("caserver.cadir"))

	s := grpc.NewServer()

	pb.RegisterWhitelistServer(s, &whitelistServer{})
	pb.RegisterCAServer(s, &CAServer{})

	// Init the crypto layer
	if err := ca.Init(); err != nil {
		slogger.Panicf("Failed initializing the crypto layer [%s]", err)
	}

	ca.CacheConfiguration()
	cap = ca.NewCA("Blockchain", ca.InitializeCommonTables)

	port := viper.GetString("caserver.port")
	if port == "" {
		slogger.Panicf("ca server port is undefined")
	}

	if sock, err := net.Listen("tcp", port); err != nil {
		log.Fatalf("failed to listen: %v", err)
		os.Exit(1)
	} else {
		s.Serve(sock)
	}
}
