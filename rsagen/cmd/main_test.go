//go:build usetempdir
// +build usetempdir

package main

import (
	"os"
	"path"
	"testing"

	"github.com/go-playground/assert"
)

func Test_app_generate(t *testing.T) {
	tmp := os.TempDir()
	size := 2048

	type fields struct {
		out  string
		bits int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "#1",
			fields: fields{
				out:  tmp,
				bits: size,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := newApp(tt.fields.out, tt.fields.bits)
			if err != nil {
				t.Errorf("newApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := a.generate(); (err != nil) != tt.wantErr {
				t.Errorf("app.generate() error = %v, wantErr %v", err, tt.wantErr)
			}

			publicKeyPath := path.Join(a.out, defaultPublicKeyName)
			checkFileSize(t, publicKeyPath)

			privatePath := path.Join(a.out, defaultPrivateKeyName)
			checkFileSize(t, privatePath)

			defer func() {
				if err := os.Remove(publicKeyPath); err != nil {
					t.Errorf("an occured error while remove public key, err: %v", err)
				}
			}()
			defer func() {
				if err := os.Remove(privatePath); err != nil {
					t.Errorf("an occured error while remove private key, err: %v", err)
				}
			}()
		})
	}
}

func checkFileSize(t *testing.T, pathToFile string) error {
	fileInfo, err := os.Stat(pathToFile)
	if err != nil {
		t.Errorf("fileInfo err: %v", err)
	}

	assert.NotEqual(t, fileInfo.Size(), 0)

	return nil
}
