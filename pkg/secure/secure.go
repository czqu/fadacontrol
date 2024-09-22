package secure

import (
	"fadacontrol/internal/base/exception"
	"golang.org/x/crypto/chacha20poly1305"
)

const MaxKeyLength = 32

type EncryptionAlgorithmEnum uint8

const (
	None                      EncryptionAlgorithmEnum = iota
	AESGCM128Algorithm                                // The AES-128GCM key is 16 bytes long
	AESGCM192Algorithm                                // The AES-192GCM key is 24 bytes long
	AESGCM256Algorithm                                // The AES-256GCM key is 32 bytes long
	ChaCha20Poly1305Algorithm                         // The ChaCha20-Poly1305 key is 32 bytes long
	Unknown                   = 0xff
)

var AESGCMAlgorithmKeyLengths = map[EncryptionAlgorithmEnum]int{

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
	switch algo {
	case AESGCM128Algorithm, AESGCM192Algorithm, AESGCM256Algorithm:
		if len(key) != AESGCMAlgorithmKeyLengths[algo] {
			return nil, exception.ErrUnknownException
		}
		return DecryptAESGCM(encryptedData, key)
	case None:
		return encryptedData, nil
	case ChaCha20Poly1305Algorithm:
		return DecryptChaCha20Poly1305(encryptedData, key)
	default:
		return nil, exception.ErrUserUnsupportedEncryptionType
	}
}
func EncryptData(algo EncryptionAlgorithmEnum, data []byte, key []byte) ([]byte, error) {
	switch algo {
	case AESGCM128Algorithm, AESGCM192Algorithm, AESGCM256Algorithm:
		if len(key) != AESGCMAlgorithmKeyLengths[algo] {
			return nil, exception.ErrUnknownException
		}
		return EncryptAESGCM(data, key)
	case None:
		return data, nil
	case ChaCha20Poly1305Algorithm:
		return EncryptChaCha20Poly1305(data, key)
	default:
		return nil, exception.ErrUserUnsupportedEncryptionType
	}
}
