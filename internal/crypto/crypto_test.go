//go:build usetempdir
// +build usetempdir

package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/go-playground/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	td := os.TempDir()
	privFile := path.Join(td, "priv.pem")
	pubFile := path.Join(td, "pub.pem")

	defer func() {
		if err := os.Remove(privFile); err != nil {
			t.Errorf("cannot remove temp private.pem file: %v", err)
		}
	}()

	defer func() {
		if err := os.Remove(pubFile); err != nil {
			t.Errorf("cannot remove temp public.pem file: %v", err)
		}
	}()

	if err := generateKeys(t, privFile, pubFile, 100); err != nil {
		t.Errorf("cannot generate keypair, err: %v", err)
	}

	type args struct {
		publicKey  string
		privateKey string
		msg        []byte
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		wantEncryptErr bool
		wantDecryptErr bool
	}{
		{
			name: "#1",
			args: args{
				publicKey:  pubFile,
				privateKey: privFile,
				msg:        []byte("s"),
			},
			wantErr:        false,
			wantEncryptErr: false,
			wantDecryptErr: false,
		},
		{
			name: "#2",
			args: args{
				publicKey:  pubFile,
				privateKey: privFile,
				msg:        []byte("this is a super secret that has a super large size"),
			},
			wantErr:        false,
			wantEncryptErr: true,
			wantDecryptErr: true,
		},
		{
			name: "#3",
			args: args{
				publicKey:  newTempFile(t),
				privateKey: newTempFile(t),
				msg:        []byte(""),
			},
			wantErr:        false,
			wantEncryptErr: true,
			wantDecryptErr: true,
		},
		{
			name: "#4",
			args: args{
				publicKey:  newTempFile(t),
				privateKey: "priv.pem",
				msg:        []byte(""),
			},
			wantErr:        true,
			wantEncryptErr: true,
			wantDecryptErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotPrivate, err := GetKeyBytes(tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				if gotPrivate == nil {
					return
				}
				t.Errorf("private key GetKeyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotEqual(t, len(gotPrivate), 0)

			gotPublic, err := GetKeyBytes(tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				if gotPublic == nil {
					return
				}
				t.Errorf("public key GetKeyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotEqual(t, len(gotPublic), 0)

			encrypted, err := Encrypt(gotPublic, tt.args.msg)
			if (err != nil) != tt.wantEncryptErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			decrypted, err := Decrypt(gotPrivate, encrypted)
			if (err != nil) != tt.wantDecryptErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantDecryptErr && !tt.wantEncryptErr {
				assert.Equal(t, string(decrypted), string(tt.args.msg))
			}
		})
	}
}

func generateKeys(t *testing.T, privFile string, pubFile string, bits int) error {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		t.Error(err)
	}

	publicKey := &privateKey.PublicKey

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("an occured error when marshal private key, err: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	err = os.WriteFile(privFile, privateKeyPEM, 0600)
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
	err = os.WriteFile(pubFile, publicKeyPEM, 0600)
	if err != nil {
		return fmt.Errorf("an occured error when write public key, err: %w", err)
	}

	return nil
}

func newTempFile(t *testing.T) string {
	t.Helper()
	td := os.TempDir()

	f, err := os.CreateTemp(td, "*")
	if err != nil {
		t.Errorf("cannot create new temp file for tests: %v", err)
	}

	if _, err := f.Write([]byte("test")); err != nil {
		t.Errorf("cannot write test in temp file for tests: %v", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("cannot close temp file for tests: %v", err)
		}
	}()

	return f.Name()
}
