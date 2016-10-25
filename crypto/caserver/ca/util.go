package ca

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"crypto/ecdsa"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	"strings"
	"crypto/rand"
	"os"
)

func VerifySignature(cert []byte, caCert []byte) error {
	c := BuildCertificateFromBytes(cert)
	caC := BuildCertificateFromBytes(caCert)

	return c.CheckSignatureFrom(caC)
}

func BuildCertificateFromBytes(cooked []byte) *x509.Certificate {
	block, _ := pem.Decode(cooked)

	if block == nil {
		return nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		caLogger.Panic(err)
	}
	return cert
}

func CreateCAKeyPair(name, path string) *ecdsa.PrivateKey {
	name = strings.Replace(name, "/", "_", -1)

	// Make sure the key pair only create once for one node
	cooked, err := ioutil.ReadFile(path + "/" + name + ".priv")
	if err == nil {
		block, _ := pem.Decode(cooked)
		if block == nil {
			return nil
		}

		pk, _ := x509.ParseECPrivateKey(block.Bytes)
		return pk
	}

	if _, err := os.Stat(path); err != nil {
		caLogger.Info("Fresh start; creating certificates keystore")

		if err := os.MkdirAll(path, 0755); err != nil {
			caLogger.Panic(err)
		}
	}

	caLogger.Debugf("Creating CA key pair. name = %s", name)  // name = Geth/v1.4.9-stable/darwin/go1.7.1

	curve := primitives.GetDefaultCurve()
	//curve := secp256k1.S256()
	//curve := elliptic.P256()

	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err == nil {
		raw, _ := x509.MarshalECPrivateKey(priv)
		cooked := pem.EncodeToMemory(
			&pem.Block{
				Type:  "ECDSA PRIVATE KEY",
				Bytes: raw,
			})
		err = ioutil.WriteFile(path+"/"+name+".priv", cooked, 0644)
		if err != nil {
			caLogger.Panic(err)
		}

		raw, _ = x509.MarshalPKIXPublicKey(&priv.PublicKey)
		cooked = pem.EncodeToMemory(
			&pem.Block{
				Type:  "ECDSA PUBLIC KEY",
				Bytes: raw,
			})
		err = ioutil.WriteFile(path+"/"+name+".pub", cooked, 0644)
		if err != nil {
			caLogger.Panic(err)
		}
	}
	if err != nil {
		caLogger.Panic(err)
	}

	return priv
}