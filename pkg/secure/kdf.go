package secure

import (
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/sha3"
)

// GenerateBcryptKey generates a key using Bcrypt.
func GenerateBcryptKey(password string, cost int) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// GenerateArgon2IDKey generates a key using Argon2id with a provided salt.
// The time parameter specifies the number of passes over the memory and the
// memory parameter specifies the size of the memory in KiB.  The number of
// threads can be adjusted to the numbers of available CPUs. The cost parameters
// should be increased as memory latency and CPU parallelism increases.
func GenerateArgon2IDKeyOneTime64MB4Threads(password string, salt []byte, time uint32, keyLen uint32) ([]byte, error) {
	key := argon2.IDKey([]byte(password), salt, time, 64*1024, 4, keyLen)
	return key, nil
}

// GenerateScryptKey generates a key using Scrypt with a provided salt.
func GenerateScryptKey(password string, salt []byte, keyLen, N, r, p int) ([]byte, error) {
	key, err := scrypt.Key([]byte(password), salt, N, r, p, keyLen)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GeneratePBKDF2Key generates a key using PBKDF2 with a provided salt.
func GeneratePBKDF2Key(password string, salt []byte, keyLen, iter int) []byte {
	key := pbkdf2.Key([]byte(password), salt, iter, keyLen, sha256.New)
	return key
}

// GenerateHKDFKey generates a key using HKDF with a provided salt.
func GenerateHKDFKey(password string, salt []byte, info []byte, keyLen int) ([]byte, error) {
	hkdf := hkdf.New(sha3.New256, []byte(password), salt, info)
	key := make([]byte, keyLen)
	_, err := hkdf.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateSalt generates a random salt.
func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
