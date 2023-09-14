package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func GetKeyBytes(pathToKey string) ([]byte, error) {
	f, err := os.Open(pathToKey)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when opening a file with a public key, err: %w", err)
	}
	defer f.Close()

	i, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("an error occurred while retrieving file stat, err: %w", err)
	}
	buf := make([]byte, i.Size())
	f.Read(buf)

	return buf, nil
}

func Encrypt(publicKey []byte, msg []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, fmt.Errorf("encrypt PEM formatted block not found")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while parse public key, err: %w", err)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), msg)
	if err != nil {
		if errors.Is(err, rsa.ErrMessageTooLong) {
			return nil, fmt.Errorf("msg size: %d, key size: %d, err: %w", binary.Size(msg), binary.Size(publicKey), err)
		}
		return nil, fmt.Errorf("an error occurred while encrypt text, err: %w", err)
	}
	return encrypted, nil
}

func Decrypt(privateKey []byte, msg []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, fmt.Errorf("decrypt PEM formatted block not found")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while parse private key, err: %w", err)
	}
	private, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("an error occurred when check key, err: %w", err)
	}

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, private, msg)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while decrypt text, err: %w", err)
	}

	return decrypted, nil
}
