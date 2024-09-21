package secure

import (
	"crypto/rand"
	"github.com/btcsuite/btcutil/base58"
)

// GenerateRandomBase58Key GenerateRandomKey generates a random byte slice of the specified length and returns its Base58 encoded string.
func GenerateRandomBase58Key(length int) (string, error) {
	// Create a byte slice to hold the random bytes.
	key := make([]byte, length)

	// Fill the byte slice with random bytes.
	if _, err := rand.Read(key); err != nil {
		return "", err
	}

	// Encode the byte slice to a Base58 string.
	encodedKey := base58.Encode(key)

	return encodedKey, nil
}

// DecodeBase58Key decodes a Base58 encoded string back to the original byte slice.
func DecodeBase58Key(encodedKey string) ([]byte, error) {
	// Decode the Base58 encoded string to a byte slice.
	decodedKey := base58.Decode(encodedKey)

	return decodedKey, nil
}
