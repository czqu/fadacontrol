package secure

import (
	"bytes"
	"crypto/rand"
	"golang.org/x/crypto/chacha20poly1305"
	"testing"
)

// TestEncryptDecryptChaCha20Poly1305 tests the encryption and decryption functions for correctness.
func TestEncryptDecryptChaCha20Poly1305(t *testing.T) {
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("This is a test plaintext.")

	// Test encryption
	ciphertext, err := EncryptChaCha20Poly1305(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptChaCha20Poly1305 failed: %v", err)
	}

	// Ensure ciphertext is not equal to plaintext
	if bytes.Equal(plaintext, ciphertext) {
		t.Error("ciphertext should not match plaintext")
	}

	// Test decryption
	decrypted, err := DecryptChaCha20Poly1305(key, ciphertext)
	if err != nil {
		t.Fatalf("DecryptChaCha20Poly1305 failed: %v", err)
	}

	// Ensure decrypted text matches the original plaintext
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted text does not match original plaintext; got %s, want %s", decrypted, plaintext)
	}
}

// TestDecryptChaCha20Poly1305WithTamperedData tests decryption with tampered data.
func TestDecryptChaCha20Poly1305WithTamperedData(t *testing.T) {
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("This is a test plaintext.")
	ciphertext, err := EncryptChaCha20Poly1305(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptChaCha20Poly1305 failed: %v", err)
	}

	// Tamper with the ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xff

	// Attempt to decrypt tampered ciphertext
	_, err = DecryptChaCha20Poly1305(key, ciphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with tampered ciphertext, but it succeeded")
	}
}

// TestEncryptChaCha20Poly1305WithInvalidKey tests encryption with an invalid key size.
func TestEncryptChaCha20Poly1305WithInvalidKey(t *testing.T) {
	invalidKey := make([]byte, 10) // Invalid key size
	plaintext := []byte("This is a test plaintext.")

	_, err := EncryptChaCha20Poly1305(invalidKey, plaintext)
	if err == nil {
		t.Fatal("expected encryption to fail with invalid key size, but it succeeded")
	}
}

// TestDecryptChaCha20Poly1305WithInvalidCiphertext tests decryption with an invalid ciphertext.
func TestDecryptChaCha20Poly1305WithInvalidCiphertext(t *testing.T) {
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	invalidCiphertext := make([]byte, 10) // Invalid ciphertext length
	_, err := DecryptChaCha20Poly1305(key, invalidCiphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with invalid ciphertext, but it succeeded")
	}
}
