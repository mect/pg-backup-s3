package crypt

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestEncryptFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "crypt-test")
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(path.Join(tmpDir, "plaintext1"), []byte("test ok"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		path    string
		outPath string
		pass    string
	}
	tests := []struct {
		name        string
		encryptArgs args
		decryptArgs args
		wantErrEnc  bool
		wantErrDec  bool
	}{
		{
			name: "encrypt test file",
			encryptArgs: args{
				path:    path.Join(tmpDir, "plaintext1"),
				outPath: path.Join(tmpDir, "enctest1"),
				pass:    "test",
			},
			decryptArgs: args{
				path:    path.Join(tmpDir, "enctest1"),
				outPath: path.Join(tmpDir, "decrypttest1"),
				pass:    "test",
			},
		},
		{
			name: "Fail to decrypt test file with incorrect pass",
			encryptArgs: args{
				path:    path.Join(tmpDir, "plaintext1"),
				outPath: path.Join(tmpDir, "enctest1"),
				pass:    "test",
			},
			decryptArgs: args{
				path:    path.Join(tmpDir, "enctest1"),
				outPath: path.Join(tmpDir, "decrypttest1"),
				pass:    "badtest",
			},
			wantErrDec: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EncryptFile(tt.encryptArgs.path, tt.encryptArgs.outPath, tt.encryptArgs.pass); (err != nil) != tt.wantErrEnc {
				t.Errorf("EncryptFile() error = %v, wantErr %v", err, tt.wantErrEnc)
			}
			if err := DecryptFile(tt.decryptArgs.path, tt.decryptArgs.outPath, tt.decryptArgs.pass); (err != nil) != tt.wantErrDec {
				t.Errorf("DecryptFile() error = %v, wantErr %v", err, tt.wantErrDec)
			}
			input, err := ioutil.ReadFile(tt.encryptArgs.path)
			if err != nil {
				t.Fatal(err)
			}
			output, err := ioutil.ReadFile(tt.decryptArgs.outPath)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(input, output) {
				t.Errorf("EncryptFile() then DecryptFile() causes a change: input %q output: %q", input, output)
			}
		})
	}
}
