package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
	"strings"
	"time"
)

const (
	defaultCertName       = "cert.pem"
	defaultPublicKeyName  = "public.pem"
	defaultPrivateKeyName = "private.pem"
	defaultOutFlag        = "o"
	defaultBitSizeFlag    = "b"
	defaultBitsSize       = 16384
)

type app struct {
	out  string
	bits int
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("an occured error: %v", err)
	}
}

func run() error {
	var out string
	var bits int

	flag.StringVar(&out, defaultOutFlag, "", "path to private.pem and public.pem keys")
	flag.IntVar(&bits, defaultBitSizeFlag, defaultBitsSize, "bits size")

	flag.Parse()

	a, err := newApp(out, bits)
	if err != nil {
		return fmt.Errorf("an occured error when parse flag, err: %w", err)
	}
	if err := a.generate(); err != nil {
		return fmt.Errorf("an occured error while generating key pairs, err: %w", err)
	}
	return nil
}

func newApp(out string, bits int) (*app, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot getting current directory, err: %w", err)
	}

	a := &app{}
	a.out = out
	a.bits = bits

	if strings.TrimSpace(a.out) == "" {
		a.out = wd
	}

	return a, nil
}

func (a *app) generate() error {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"NoOrganization"},
			Country:      []string{"RU"},
		},
		DNSNames:     []string{"*"},
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, a.bits)
	if err != nil {
		return fmt.Errorf("an occured error when generate rsa key, err: %w", err)
	}
	publicKey := &privateKey.PublicKey
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return fmt.Errorf("an occured error when create cert, err: %w", err)
	}
	crtPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	err = os.WriteFile(path.Join(a.out, defaultCertName), crtPEM, 0600)
	if err != nil {
		return fmt.Errorf("an occured error when write public key, err: %w", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("an occured error when marshal private key, err: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	err = os.WriteFile(path.Join(a.out, defaultPrivateKeyName), privateKeyPEM, 0600)
	if err != nil {
		return fmt.Errorf("an occured error when write private key, err: %w", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("an occured error when marshal public key, err: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(path.Join(a.out, defaultPublicKeyName), publicKeyPEM, 0600)
	if err != nil {
		return fmt.Errorf("an occured error when write public key, err: %w", err)
	}

	return nil
}
