package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

// EncryptAESGCM encrypts plaintext using AES-GCM with the provided key.
func EncryptAESGCM(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	// Concatenate nonce and ciphertext to preserve nonce for decryption
	result := append(nonce, ciphertext...)
	return result, nil
}

// DecryptAESGCM decrypts ciphertext using AES-GCM with the provided key.
func DecryptAESGCM(key []byte, encrypted []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(encrypted) < aesGCM.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := encrypted[:aesGCM.NonceSize()]
	ciphertext := encrypted[aesGCM.NonceSize():]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
