package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/alxarno/swap/settings"
)

const (
	createCertificateFunc = "createCertificate "
	initCertificateFunc   = "InitCert"
	certPath              = "./swap.crt"
	keyPath               = "./swap.key"
)

var (
	Cert *tls.Certificate = nil
)

func InitCert() {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		createCertificate()
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		createCertificate()
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatalf("%s -> failed load key and cert -> %s", initCertificateFunc, err.Error())
	}
	Cert = &cert
}

func createCertificate() {
	priv, err := rsa.GenerateKey(rand.Reader, settings.ServiceSettings.Cert.RsaBits)
	if err != nil {
		log.Fatalf("%s -> failed to generate private key: %s", createCertificateFunc, err.Error())
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("%s -> failed to generate serial number: %s", createCertificateFunc, err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{settings.ServiceSettings.Cert.Org},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range settings.ServiceSettings.Cert.Hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	template.IsCA = false

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("%s -> failder to create certificate: %s", createCertificateFunc, err.Error())
	}

	certOut, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("%s -> failed to open cert.pem for writing: %s", createCertificateFunc, err.Error())
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("%s -> failed to write data to cert: %s", createCertificateFunc, err.Error())
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("%s -> error closing cert: %s", createCertificateFunc, err.Error())
	}
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("%s -> failed to open key for writing: %s", createCertificateFunc, err.Error())
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		log.Fatalf("%s -> failed to write data to key: %s", createCertificateFunc, err.Error())
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("%s ->error closing key: %s", createCertificateFunc, err.Error())
	}
	log.Println(fmt.Sprintf("New key - %s and certificate - %s created", keyPath, certPath))
}
