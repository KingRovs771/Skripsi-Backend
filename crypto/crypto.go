package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	"strings"
)

// getEncryptionKey mengembalikan 32-byte key untuk AES-256.
func getEncryptionKey() []byte {
	var keyRaw string

	if keyEnv := os.Getenv("SINDAS_ENCRYPTION_KEY"); keyEnv != "" {
		keyRaw = keyEnv
	} else {
		fileBytes, err := os.ReadFile("/etc/sindas/encryption_key")
		if err == nil && len(fileBytes) > 0 {
			keyRaw = strings.TrimSpace(string(fileBytes))
		}
	}

	if keyRaw == "" {
		keyRaw = "SINDAS_DEFAULT_SECURE_KEY_2026_FALLBACK"
	}

	hasher := sha256.New()
	hasher.Write([]byte(keyRaw))
	return hasher.Sum(nil)
}

// EncryptField mengenkripsi plaintext menggunakan AES-256-GCM.
func EncryptField(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nil, nonce, []byte(plainText), nil)
	combined := append(nonce, cipherText...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptField mendekripsi cipherText (Base64) kembali menjadi plaintext.
func DecryptField(cipherTextStr string) (string, error) {
	if cipherTextStr == "" {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cipherTextStr)
	if err != nil {
		return cipherTextStr, nil
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return cipherTextStr, nil
	}

	nonce := decoded[:nonceSize]
	actualCipherText := decoded[nonceSize:]

	plainTextBytes, err := gcm.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return cipherTextStr, nil
	}

	return string(plainTextBytes), nil
}
