package tokens

import (
	"crypto/rand"
	"encoding/base64"

	errorMessages "github.com/denwong47/pigeon-hole/pkg/errors"
)

const TokenLength = 64

// GenerateToken generates a random token of a given length.
// The token shall bear no meaning and be cryptographically secure.
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", errorMessages.ErrTokenGeneration
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
