package secure

import (
	"crypto/sha256"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/sha3"
	"testing"
)

func TestGenerateBcryptKey(t *testing.T) {
	password := "testPassword"
	cost := 12

	key, err := GenerateBcryptKey(password, cost)
	if err != nil {
		t.Fatalf("GenerateBcryptKey failed: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(key, []byte(password)); err != nil {
		t.Fatalf("Bcrypt key verification failed: %v", err)
	}
}

func TestGenerateArgon2IDKey(t *testing.T) {
	password := "testPassword"
	salt, err := GenerateSalt(16)
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	time := uint32(1)
	keyLen := 32

	key, err := GenerateArgon2IDKeyOneTime64MB4Threads(password, salt, time, uint32(keyLen))
	if err != nil {
		t.Fatalf("GenerateArgon2IDKey failed: %v", err)
	}

	expectedKey := argon2.IDKey([]byte(password), salt, time, 64*1024, 4, uint32(keyLen))
	if !equal(key, expectedKey) {
		t.Fatalf("Argon2ID key mismatch: got %x, want %x", key, expectedKey)
	}
}

func TestGenerateScryptKey(t *testing.T) {
	password := "testPassword"
	salt, err := GenerateSalt(16)
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	keyLen := 32
	N := 1 << 14
	r := 8
	p := 1

	key, err := GenerateScryptKey(password, salt, keyLen, N, r, p)
	if err != nil {
		t.Fatalf("GenerateScryptKey failed: %v", err)
	}

	expectedKey, err := scrypt.Key([]byte(password), salt, N, r, p, keyLen)
	if err != nil {
		t.Fatalf("Scrypt key generation failed: %v", err)
	}

	if !equal(key, expectedKey) {
		t.Fatalf("Scrypt key mismatch: got %x, want %x", key, expectedKey)
	}
}

func TestGeneratePBKDF2Key(t *testing.T) {
	password := "testPassword"
	salt, err := GenerateSalt(16)
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	keyLen := 32
	iter := 10000

	key := GeneratePBKDF2Key(password, salt, keyLen, iter)
	expectedKey := pbkdf2.Key([]byte(password), salt, iter, keyLen, sha256.New)

	if !equal(key, expectedKey) {
		t.Fatalf("PBKDF2 key mismatch: got %x, want %x", key, expectedKey)
	}
}

func TestGenerateHKDFKey(t *testing.T) {
	password := "testPassword"
	salt, err := GenerateSalt(16)
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	info := []byte("info")
	keyLen := 32

	key, err := GenerateHKDFKey(password, salt, info, keyLen)
	if err != nil {
		t.Fatalf("GenerateHKDFKey failed: %v", err)
	}

	hkdf := hkdf.New(sha3.New256, []byte(password), salt, info)
	expectedKey := make([]byte, keyLen)
	_, err = hkdf.Read(expectedKey)
	if err != nil {
		t.Fatalf("HKDF key generation failed: %v", err)
	}

	if !equal(key, expectedKey) {
		t.Fatalf("HKDF key mismatch: got %x, want %x", key, expectedKey)
	}
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
