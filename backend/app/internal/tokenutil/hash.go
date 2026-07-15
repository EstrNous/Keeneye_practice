package tokenutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func NewToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	hash = Hash(raw)
	return raw, hash, nil
}

func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
