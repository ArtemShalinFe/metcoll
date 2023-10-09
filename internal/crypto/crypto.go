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

	"go.uber.org/zap"
)

func GetKeyBytes(pathToKey string) ([]byte, error) {
	f, err := os.Open(pathToKey)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("an error occurred when opening a file with a key, err: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			zap.S().Errorf("an error occurred when closing file %s, err: %w", pathToKey, err)
		}
	}()

	i, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("an error occurred while retrieving file %s, err: %w", pathToKey, err)
	}
	buf := make([]byte, i.Size())

	if _, err := f.Read(buf); err != nil {
		return nil, fmt.Errorf("an error occurred when reading file %s, err: %w", pathToKey, err)
	}

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
