package distributed

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256sum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return []byte(hex.EncodeToString(hash[:]))
}
