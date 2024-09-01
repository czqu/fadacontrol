package secure

import (
	"bytes"
	"crypto/rand"
	"testing"
)

// TestEncryptDecryptAESGCM tests the encryption and decryption functions for correctness.
func TestEncryptDecryptAESGCM(t *testing.T) {
	key := make([]byte, 32) // AES-256 key size
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("This is a test plaintext.")

	// Test encryption
	ciphertext, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptAESGCM failed: %v", err)
	}

	// Ensure ciphertext is not equal to plaintext
	if bytes.Equal(plaintext, ciphertext) {
		t.Error("ciphertext should not match plaintext")
	}

	// Test decryption
	decrypted, err := DecryptAESGCM(key, ciphertext)
	if err != nil {
		t.Fatalf("DecryptAESGCM failed: %v", err)
	}

	// Ensure decrypted text matches the original plaintext
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted text does not match original plaintext; got %s, want %s", decrypted, plaintext)
	}
}

// TestDecryptAESGCMWithTamperedData tests decryption with tampered ciphertext.
func TestDecryptAESGCMWithTamperedData(t *testing.T) {
	key := make([]byte, 32) // AES-256 key size
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("This is a test plaintext.")
	ciphertext, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptAESGCM failed: %v", err)
	}

	// Tamper with the ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xff

	// Attempt to decrypt tampered ciphertext
	_, err = DecryptAESGCM(key, ciphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with tampered ciphertext, but it succeeded")
	}
}

// TestEncryptAESGCMWithInvalidKey tests encryption with an invalid key size.
func TestEncryptAESGCMWithInvalidKey(t *testing.T) {
	invalidKey := make([]byte, 15) // Invalid key size for AES
	plaintext := []byte("This is a test plaintext.")

	_, err := EncryptAESGCM(invalidKey, plaintext)
	if err == nil {
		t.Fatal("expected encryption to fail with invalid key size, but it succeeded")
	}
}

// TestDecryptAESGCMWithInvalidCiphertext tests decryption with an invalid ciphertext.
func TestDecryptAESGCMWithInvalidCiphertext(t *testing.T) {
	key := make([]byte, 32) // AES-256 key size
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	invalidCiphertext := make([]byte, 10) // Invalid ciphertext length
	_, err := DecryptAESGCM(key, invalidCiphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with invalid ciphertext, but it succeeded")
	}
}
