package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type CursorCrypto struct {
	cipher cipher.AEAD
}

func NewCursorCrypto(secretKey string) (*CursorCrypto, error) {
	secret := Hash256(secretKey)
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &CursorCrypto{cipher: gcm}, nil
}

func (cc *CursorCrypto) Encrypt(value string) (string, error) {
	nonce := make([]byte, cc.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := cc.cipher.Seal(nil, nonce, []byte(value), nil)
	nonce = append(nonce, encrypted...)
	return base64.URLEncoding.EncodeToString(nonce), nil
}

func (cc *CursorCrypto) Decrypt(cursor string) (string, error) {
	if cursor == "" {
		return "", nil
	}

	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", fmt.Errorf("invalid cursor format")
	}

	nonceSize := cc.cipher.NonceSize()
	if len(decoded) < nonceSize {
		return "", fmt.Errorf("invalid cursor size")
	}

	nonce := decoded[:nonceSize]
	ciphertext := decoded[nonceSize:]

	plaintext, err := cc.cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("cursor decryption failed")
	}

	return string(plaintext), nil
}
