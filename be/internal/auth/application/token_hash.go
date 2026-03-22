package application

import (
	"crypto/sha256"
	"encoding/hex"
)

func tokenHash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
