package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const apiKeyPrefix = "mrk_"

// GenerateAPIKey creates a new API key and returns (plaintext, hash, displayPrefix).
func GenerateAPIKey() (string, string, string) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	plain := apiKeyPrefix + hex.EncodeToString(b)
	hash := HashAPIKey(plain)
	prefix := plain[:12] + "..."
	return plain, hash, prefix
}

// HashAPIKey returns a SHA-256 hex digest of the API key.
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
