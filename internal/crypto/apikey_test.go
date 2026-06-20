package crypto

import (
	"strings"
	"testing"
)

func TestEncryptAPIKeyRoundTrip(t *testing.T) {
	ciphertext, err := EncryptAPIKey("secret-one", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}
	if !strings.HasPrefix(ciphertext, encryptedAPIKeyPrefix) {
		t.Fatalf("ciphertext prefix = %q, want %q", ciphertext, encryptedAPIKeyPrefix)
	}
	if strings.Contains(ciphertext, "secret-one") {
		t.Fatalf("ciphertext includes plaintext: %q", ciphertext)
	}
	plaintext, err := DecryptAPIKey(ciphertext, "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}
	if plaintext != "secret-one" {
		t.Fatalf("plaintext = %q, want secret-one", plaintext)
	}
}

func TestEncryptAPIKeyUsesRandomNonce(t *testing.T) {
	first, err := EncryptAPIKey("same-secret", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("EncryptAPIKey(first) error = %v", err)
	}
	second, err := EncryptAPIKey("same-secret", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("EncryptAPIKey(second) error = %v", err)
	}
	if first == second {
		t.Fatal("EncryptAPIKey produced identical ciphertexts for the same plaintext")
	}
}

func TestDecryptAPIKeyTreatsUnprefixedValuesAsLegacyPlaintext(t *testing.T) {
	plaintext, err := DecryptAPIKey("legacy-plaintext", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}
	if plaintext != "legacy-plaintext" {
		t.Fatalf("plaintext = %q, want legacy-plaintext", plaintext)
	}
}

func TestDecryptAPIKeyRejectsWrongSecret(t *testing.T) {
	ciphertext, err := EncryptAPIKey("secret-one", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}
	if _, err := DecryptAPIKey(ciphertext, "different-jwt-secret-32-bytes-value"); err == nil {
		t.Fatal("DecryptAPIKey() error = nil, want authentication failure")
	}
}

func TestEncryptAPIKeyFallsBackForEmptySecret(t *testing.T) {
	ciphertext, err := EncryptAPIKey("test-secret", "")
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}
	plaintext, err := DecryptAPIKey(ciphertext, "")
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}
	if plaintext != "test-secret" {
		t.Fatalf("plaintext = %q, want test-secret", plaintext)
	}
}
