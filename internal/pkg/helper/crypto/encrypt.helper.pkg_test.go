package crypto

import (
	"strings"
	"testing"

	"vibe-ddd-golang/internal/config"
)

func cryptoConfig(encryptKey, ivKey, cipher string) config.CryptoConfig {
	return config.CryptoConfig{
		EncryptKey: encryptKey,
		IVKey:      ivKey,
		Cipher:     cipher,
	}
}

func TestEncryptDecryptAESCBC_WithCIPHER(t *testing.T) {
	cfg := cryptoConfig("12345678901234567890123456789012", "1234567890123456", "aes-256-cbc")

	plain := "hello-world"
	enc, err := EncryptAESCBC(cfg, plain)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	dec, err := DecryptAESCBC(cfg, enc)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if dec != plain {
		t.Fatalf("expected %s got %s", plain, dec)
	}
}

func TestEncryptDecryptAESCBC_DefaultCipher(t *testing.T) {
	cfg := cryptoConfig("12345678901234567890123456789012", "1234567890123456", "")

	plain := "default-cipher"
	enc, err := EncryptAESCBC(cfg, plain)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	dec, err := DecryptAESCBC(cfg, enc)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if dec != plain {
		t.Fatalf("expected %s got %s", plain, dec)
	}
}

func TestEncryptAESCBC_UnsupportedCipher(t *testing.T) {
	cfg := cryptoConfig("12345678901234567890123456789012", "1234567890123456", "aes-128-cbc")
	_, err := EncryptAESCBC(cfg, "x")
	if err == nil || !strings.Contains(err.Error(), "unsupported cipher") {
		t.Fatalf("expected unsupported cipher error got %v", err)
	}
}

func TestEncryptAESCBC_InvalidKeyIvLength(t *testing.T) {
	cfg := cryptoConfig("shortkey", "shortiv", "")
	_, err := EncryptAESCBC(cfg, "x")
	if err == nil || !strings.Contains(err.Error(), "ENCRYPT_KEY must be 32 bytes") {
		t.Fatalf("expected key/iv length error got %v", err)
	}
}
