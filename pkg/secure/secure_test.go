package secure

import (
	"bytes"
	"crypto/rand"
	"fadacontrol/internal/base/exception"
	"testing"
)

// TestEncryptDecrypt tests the encryption and decryption functions for various algorithms and key lengths.
func TestEncryptDecrypt(t *testing.T) {
	data := []byte("This is a test data")

	tests := []struct {
		algo EncryptionAlgorithmEnum
		key  []byte
	}{
		{AESGCM128Algorithm, generateRandomKey(AlgorithmKeyLengths[AESGCM128Algorithm])},
		{AESGCM192Algorithm, generateRandomKey(AlgorithmKeyLengths[AESGCM192Algorithm])},
		{AESGCM256Algorithm, generateRandomKey(AlgorithmKeyLengths[AESGCM256Algorithm])},
		{ChaCha20Poly1305Algorithm, generateRandomKey(AlgorithmKeyLengths[ChaCha20Poly1305Algorithm])},
		{None, generateRandomKey(AlgorithmKeyLengths[None])},
		{AESGCM128Algorithm, generateRandomKey(35)},
		{AESGCM192Algorithm, generateRandomKey(35)},
		{AESGCM256Algorithm, generateRandomKey(35)},
		{ChaCha20Poly1305Algorithm, generateRandomKey(35)},
		{None, generateRandomKey(35)},
	}

	for _, tt := range tests {
		encryptedData, err := EncryptData(tt.algo, data, tt.key)
		if err != nil {
			t.Fatalf("Encryption failed for algorithm %v: %v", tt.algo, err)
		}

		decryptedData, err := DecryptData(tt.algo, encryptedData, tt.key)
		if err != nil {
			t.Fatalf("Decryption failed for algorithm %v: %v", tt.algo, err)
		}

		if !bytes.Equal(data, decryptedData) {
			t.Errorf("Decrypted data does not match original data for algorithm %v", tt.algo)
		}
	}
}

// TestInvalidKeyLength tests the encryption and decryption functions with invalid key lengths.
func TestInvalidKeyLength(t *testing.T) {
	data := []byte("This is a test data")

	tests := []struct {
		algo EncryptionAlgorithmEnum
		key  []byte
	}{
		{AESGCM128Algorithm, generateRandomKey(AlgorithmKeyLengths[AESGCM128Algorithm] - 1)},
		{AESGCM192Algorithm, generateRandomKey(AlgorithmKeyLengths[AESGCM192Algorithm] - 1)},
		{AESGCM256Algorithm, generateRandomKey(AlgorithmKeyLengths[AESGCM256Algorithm] - 1)},
		{ChaCha20Poly1305Algorithm, generateRandomKey(AlgorithmKeyLengths[ChaCha20Poly1305Algorithm] - 1)},
	}

	for _, tt := range tests {
		_, err := EncryptData(tt.algo, data, tt.key)
		if err != exception.ErrUserInvalidAlgoKeyLen {
			t.Errorf("Expected ErrUserInvalidAlgoKeyLen for algorithm %v with short key, got %v", tt.algo, err)
		}

		_, err = DecryptData(tt.algo, data, tt.key)
		if err != exception.ErrUserInvalidAlgoKeyLen {
			t.Errorf("Expected ErrUserInvalidAlgoKeyLen for algorithm %v with short key, got %v", tt.algo, err)
		}
	}
}

// generateRandomKey generates a random key of the specified length.
func generateRandomKey(length int) []byte {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		panic("Failed to generate random key")
	}
	return key
}
