package crypto

import (
	"strings"
	"testing"
)

func TestCursorCrypto_EncryptDecrypt(t *testing.T) {
	secret := "my-secret-key"
	crypto, err := NewCursorCrypto(secret)
	if err != nil {
		t.Fatalf("NewCursorCrypto() error = %v", err)
	}

	original := "playground-data-123"
	ciphertext, err := crypto.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	plaintext, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if plaintext != original {
		t.Errorf("Decrypt() = %q, want %q", plaintext, original)
	}
}

func TestCursorCrypto_EmptyString(t *testing.T) {
	crypto, err := NewCursorCrypto("key")
	if err != nil {
		t.Fatalf("NewCursorCrypto() error = %v", err)
	}
	ciphertext, err := crypto.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	plaintext, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "" {
		t.Errorf("Decrypt() = %q, want empty string", plaintext)
	}
}

func TestCursorCrypto_Decrypt_InvalidCursor(t *testing.T) {
	crypto, err := NewCursorCrypto("key")
	if err != nil {
		t.Fatalf("NewCursorCrypto() error = %v", err)
	}
	_, err = crypto.Decrypt("not-base64!!")
	if err == nil || !strings.Contains(err.Error(), "invalid cursor format") {
		t.Errorf("Decrypt() error = %v, want error containing 'invalid cursor format'", err)
	}
}

func TestCursorCrypto_Decrypt_InvalidSize(t *testing.T) {
	crypto, err := NewCursorCrypto("key")
	if err != nil {
		t.Fatalf("NewCursorCrypto() error = %v", err)
	}
	// base64 of 3 bytes, which is less than nonce size
	invalid := "QUJD"
	_, err = crypto.Decrypt(invalid)
	if err == nil || !strings.Contains(err.Error(), "invalid cursor size") {
		t.Errorf("Decrypt() error = %v, want error containing 'invalid cursor size'", err)
	}
}

func TestCursorCrypto_Decrypt_WrongKey(t *testing.T) {
	crypto1, err := NewCursorCrypto("key1")
	if err != nil {
		t.Fatalf("NewCursorCrypto() error = %v", err)
	}
	crypto2, err := NewCursorCrypto("key2")
	if err != nil {
		t.Fatalf("NewCursorCrypto() error = %v", err)
	}
	ciphertext, err := crypto1.Encrypt("data")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	_, err = crypto2.Decrypt(ciphertext)
	if err == nil || !strings.Contains(err.Error(), "cursor decryption failed") {
		t.Errorf("Decrypt() error = %v, want error containing 'cursor decryption failed'", err)
	}
}
