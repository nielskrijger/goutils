package goutils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/mr-tron/base58"
	uuid "github.com/satori/go.uuid" // nolint
)

// GenerateShortID returns a base58-encoded UUID which is 22
// characters long.
func GenerateShortID() string {
	return base58.Encode(uuid.NewV4().Bytes())
}

// GenerateRandomBytes returns securely generated random bytes.
// It returns an error if the system's secure random number
// generator fails.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("reading random bytes: %w", err)
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// random string.
//
// It returns an error if the system's secure random number
// generator fails.
func GenerateRandomString(bytes int) (string, error) {
	b, err := GenerateRandomBytes(bytes)

	return base64.URLEncoding.EncodeToString(b), err
}
