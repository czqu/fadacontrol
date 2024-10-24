package secure

import (
	"fadacontrol/internal/base/exception"
	"golang.org/x/crypto/chacha20poly1305"
)

const MaxKeyLength = 32

type EncryptionAlgorithmEnum uint8

const (
	NoEncryption              EncryptionAlgorithmEnum = iota
	AESGCM128Algorithm                                // The AES-128GCM key is 16 bytes long
	AESGCM192Algorithm                                // The AES-192GCM key is 24 bytes long
	AESGCM256Algorithm                                // The AES-256GCM key is 32 bytes long
	ChaCha20Poly1305Algorithm                         // The ChaCha20-Poly1305 key is 32 bytes long
	Unknown                   = 0xff
)

var AlgorithmKeyLengths = map[EncryptionAlgorithmEnum]int{
	NoEncryption: 0,

	AESGCM128Algorithm:        16, // 128-bit AES-GCM key length
	AESGCM192Algorithm:        24, // 192-bit AES-GCM key length
	AESGCM256Algorithm:        32, // 256-bit AES-GCM key length
	ChaCha20Poly1305Algorithm: chacha20poly1305.KeySize,
}
var AlgorithmNames = map[EncryptionAlgorithmEnum]string{
	AESGCM128Algorithm:        "AES-GCM128",
	AESGCM192Algorithm:        "AES-GCM192",
	AESGCM256Algorithm:        "AES-GCM256",
	ChaCha20Poly1305Algorithm: "ChaCha20Poly1305",
}

func DecryptData(algo EncryptionAlgorithmEnum, encryptedData []byte, key []byte) ([]byte, error) {
	err := checkAlgoKeyLen(algo, key)
	if err != nil {
		return nil, err
	}
	key = key[:AlgorithmKeyLengths[algo]]
	switch algo {
	case AESGCM128Algorithm, AESGCM192Algorithm, AESGCM256Algorithm:

		return DecryptAESGCM(key, encryptedData)
	case NoEncryption:
		return encryptedData, nil
	case ChaCha20Poly1305Algorithm:
		return DecryptChaCha20Poly1305(key, encryptedData)
	default:
		return nil, exception.ErrUserUnsupportedEncryptionType
	}
}
func EncryptData(algo EncryptionAlgorithmEnum, data []byte, key []byte) ([]byte, error) {
	err := checkAlgoKeyLen(algo, key)
	if err != nil {
		return nil, err
	}
	key = key[:AlgorithmKeyLengths[algo]]
	switch algo {
	case AESGCM128Algorithm, AESGCM192Algorithm, AESGCM256Algorithm:

		return EncryptAESGCM(key, data)
	case NoEncryption:
		return data, nil
	case ChaCha20Poly1305Algorithm:
		return EncryptChaCha20Poly1305(key, data)
	default:
		return nil, exception.ErrUserUnsupportedEncryptionType
	}
}
func checkAlgoKeyLen(algo EncryptionAlgorithmEnum, key []byte) error {
	if len(key) < AlgorithmKeyLengths[algo] {
		return exception.ErrSystemInvalidAlgoKeyLen
	}
	return nil
}
