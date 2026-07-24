package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"vibe-ddd-golang/internal/config"
)

func EncryptDips(cfg config.CryptoConfig, text string) (string, error) {
	if opensslVersion := "1.1.1"; opensslVersion <= "1.0.1f" {
		return "", errors.New("OpenSSL Version too old, vulnerability to Heartbleed")
	}

	ivLength := cfg.DIPSIVLength
	if ivLength == 0 {
		ivLength = 16
	}

	iv := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	key, method := []byte(cfg.DIPSPassword), cfg.DIPSAESMethod
	if block, err := aes.NewCipher(key); err != nil || method != "aes-256-cbc" {
		return "", fmt.Errorf("failed to create cipher block or unsupported method: %w", err)
	} else {
		padding := aes.BlockSize - len(text)%aes.BlockSize
		paddedText := append([]byte(text), bytes.Repeat([]byte{byte(padding)}, padding)...)

		encrypted := make([]byte, len(paddedText))
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, paddedText)

		return fmt.Sprintf("%s:%s", hex.EncodeToString(iv), hex.EncodeToString(encrypted)), nil
	}
}

func DecryptDips(cfg config.CryptoConfig, text string) (string, error) {
	parts := strings.SplitN(text, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid input text")
	}

	iv, err := hex.DecodeString(parts[0])
	encryptedText, err2 := hex.DecodeString(parts[1])
	if err != nil || err2 != nil {
		return "", fmt.Errorf("failed to decode IV or encrypted text: %w", err)
	}

	key := []byte(cfg.DIPSPassword)
	if block, err := aes.NewCipher(key); err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	} else {
		decrypted := make([]byte, len(encryptedText))
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(decrypted, encryptedText)

		unpadding := int(decrypted[len(decrypted)-1])
		return string(decrypted[:len(decrypted)-unpadding]), nil
	}
}

func HMACSHA256(cfg config.CryptoConfig, str string) (string, error) {
	key := cfg.EncryptKey
	if key == "" {
		return "", errors.New("ENCRYPT_KEY environment variable is not set")
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func HMACSHA1(str, key string) (string, error) {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(str))
	data := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(data), nil
}

func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return nil
	}
	padSize := int(data[length-1])
	if padSize > length || padSize == 0 {
		return nil
	}
	for _, b := range data[length-padSize:] {
		if int(b) != padSize {
			return nil
		}
	}
	return data[:length-padSize]
}

func resolveCipher(cfg config.CryptoConfig) (string, error) {
	c := cfg.Cipher
	if c == "" {
		c = "aes-256-cbc" // default
	}
	if c != "aes-256-cbc" {
		return "", fmt.Errorf("unsupported cipher: %s (only aes-256-cbc is supported)", c)
	}
	return c, nil
}

func EncryptAESCBC(cfg config.CryptoConfig, val string) (string, error) {
	if _, err := resolveCipher(cfg); err != nil {
		return "", err
	}

	key := cfg.EncryptKey
	iv := cfg.IVKey
	if len(key) != 32 || len(iv) != 16 {
		return "", errors.New("ENCRYPT_KEY must be 32 bytes and IV_KEY must be 16 bytes")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	paddedVal := pad([]byte(val), block.BlockSize())
	ciphertext := make([]byte, len(paddedVal))

	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, paddedVal)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptAESCBC(cfg config.CryptoConfig, encryptedVal string) (string, error) {
	if _, err := resolveCipher(cfg); err != nil {
		return "", err
	}

	key := cfg.EncryptKey
	iv := cfg.IVKey
	if len(key) != 32 || len(iv) != 16 {
		return "", errors.New("ENCRYPT_KEY must be 32 bytes and IV_KEY must be 16 bytes")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedVal)
	if err != nil {
		return "", err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(ciphertext))

	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext = unpad(plaintext)
	if plaintext == nil {
		return "", errors.New("padding error")
	}

	return string(plaintext), nil
}
