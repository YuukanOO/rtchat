package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// GenerateUID generates a nonce of the given size.
func GenerateUID(size int) string {
	nonceBytes := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, nonceBytes)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(nonceBytes)
}
