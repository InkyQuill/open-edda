package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	encryptedAPIKeyPrefix = "edda:v1:"
	encryptionLabel       = "open-edda-api-key-encryption"
	devTestSecret         = "test-encryption-key-for-edda-dev-only-32bytes"
)

var ErrInvalidCiphertext = errors.New("invalid encrypted API key")

func EncryptAPIKey(plaintext, jwtSecret string) (string, error) {
	gcm, err := apiKeyGCM(jwtSecret)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(cryptorand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate API key nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encryptedAPIKeyPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func DecryptAPIKey(value, jwtSecret string) (string, error) {
	if !strings.HasPrefix(value, encryptedAPIKeyPrefix) {
		return value, nil
	}
	encoded := strings.TrimPrefix(value, encryptedAPIKeyPrefix)
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("%w: decode ciphertext: %w", ErrInvalidCiphertext, err)
	}
	gcm, err := apiKeyGCM(jwtSecret)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", ErrInvalidCiphertext
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: authentication failed", ErrInvalidCiphertext)
	}
	return string(plaintext), nil
}

func apiKeyGCM(jwtSecret string) (cipher.AEAD, error) {
	key := deriveAPIKeyEncryptionKey(jwtSecret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create API key cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create API key GCM: %w", err)
	}
	return gcm, nil
}

func deriveAPIKeyEncryptionKey(jwtSecret string) [32]byte {
	if jwtSecret == "" {
		jwtSecret = devTestSecret
	}
	return sha256.Sum256([]byte(encryptionLabel + jwtSecret))
}
