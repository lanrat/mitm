package server

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"strings"
	"time"
)

// from: https://golang.org/src/crypto/tls/generate_cert.go

func getTLSConfig() *tls.Config {
	cfg := &tls.Config{}
	cfg.GetCertificate = getCertificateHook
	ca := readCA()
	cert := mkCert("*", ca)
	// TODO make certs generated on the fly for each request
	cfg.Certificates = append(cfg.Certificates, cert)
	cfg.BuildNameToCertificate()
	return cfg
}

func getCertificateHook(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	log.Printf("Got TLS connection for [%s] %q", helloInfo.Conn.RemoteAddr(), helloInfo.ServerName)
	return nil, nil
}

func mkCert(host string, ca *x509.Certificate) tls.Certificate {
	//host := "*"
	var priv interface{}
	var err error
	priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}
	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}
	notBefore := time.Now().Add(time.Hour * -24)
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// caCert := &template
	// if ca != nil {
	// 	log.Printf("Using CA %v", ca)
	// 	caCert = ca.Leaf
	// 	log.Printf("leaf %v", caCert)
	// }
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, ca, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}
	pemDerBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	pemPrivBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	cert, err := tls.X509KeyPair(pemDerBytes, pemPrivBytes)
	if err != nil {
		log.Fatalf("Unable to make cert: %v", err)
	}
	return cert
}

const caKeyPath = "http/ca.key"
const caCertPath = "http/ca.pem"

func readCA() *x509.Certificate {
	b1, err := ioutil.ReadFile(caKeyPath)
	check(err)
	b2, err := ioutil.ReadFile(caCertPath)
	check(err)
	// block, _ := pem.Decode(b)
	// if block == nil {
	// 	log.Fatal("pem.decode was nil for CA")
	// // }
	// privKey, err := x509.ParsePKCS1PrivateKey(b1)
	// check(err)
	// pubKey, err := x509.ParsePKCS1PublicKey(b2)
	// check(err)
	ca, err := tls.X509KeyPair(b2, b1)
	check(err)
	caCert, err := x509.ParseCertificate(ca.Certificate[0])
	check(err)
	return caCert
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}
