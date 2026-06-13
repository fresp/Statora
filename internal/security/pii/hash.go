package pii

import (
	"crypto/sha256"
	"encoding/hex"
)

func Hash(normalized string) string {
	h := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(h[:])
}
