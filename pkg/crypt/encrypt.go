package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func EncryptFile(path, outPath, pass string) error {
	plaintext, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	h := sha256.New()
	key := h.Sum([]byte(pass))

	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// create a new file for saving the encrypted data.
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(ciphertext))
	if err != nil {
		return err
	}

	return nil
}

func DecryptFile(path, outPath, pass string) error {
	ciphertext, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	h := sha256.New()
	key := h.Sum([]byte(pass))

	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create gcm: %w", err)
	}
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	// create a new file for saving the encrypted data.
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	_, err = io.Copy(f, bytes.NewReader(plaintext))
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
