package secure

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
)

type SignAlgorithm string

const (
	None        = "None"
	UNKNOWN     = "Unknown"
	HMAC_SHA256 = "HmacSHA256"
	HMAC_SHA1   = "HmacSHA1"
	HMAC_MD5    = "HmacMD5"
	HMAC_SHA512 = "HmacSHA512"
)

func CalculateHMAC(data, secret string, algo SignAlgorithm) (string, error) {
	var hmacHash hash.Hash

	// Choose the algorithm based on input
	switch algo {
	case HMAC_SHA256:
		hmacHash = hmac.New(sha256.New, []byte(secret))
	case HMAC_SHA1:
		hmacHash = hmac.New(sha1.New, []byte(secret))
	case HMAC_MD5:
		hmacHash = hmac.New(md5.New, []byte(secret))
	case HMAC_SHA512:
		hmacHash = hmac.New(sha512.New, []byte(secret))
	default:
		return "", errors.New("unsupported algorithm")
	}

	// Write the data to the HMAC hash
	_, err := hmacHash.Write([]byte(data))
	if err != nil {
		return "", err
	}

	// Get the final hash result
	hash := hmacHash.Sum(nil)

	// Return the base64 encoded hash
	return hex.EncodeToString(hash), nil
}
func ValidateHMAC(data, secret, signature string, algo SignAlgorithm) bool {
	if algo == None {
		return true
	}
	ret, err := CalculateHMAC(data, secret, algo)
	if err != nil {
		return false
	}
	return ret == signature

}
