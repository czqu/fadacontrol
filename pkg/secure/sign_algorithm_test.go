package secure

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestCalculateHMAC(t *testing.T) {
	data := "test data"
	secret := "test secret"

	// Compute expected values
	hmacSHA256 := hmac.New(sha256.New, []byte(secret))
	hmacSHA256.Write([]byte(data))
	expectedSHA256 := base64.StdEncoding.EncodeToString(hmacSHA256.Sum(nil))

	hmacSHA1 := hmac.New(sha1.New, []byte(secret))
	hmacSHA1.Write([]byte(data))
	expectedSHA1 := base64.StdEncoding.EncodeToString(hmacSHA1.Sum(nil))

	hmacMD5 := hmac.New(md5.New, []byte(secret))
	hmacMD5.Write([]byte(data))
	expectedMD5 := base64.StdEncoding.EncodeToString(hmacMD5.Sum(nil))

	// Test HMAC_SHA256
	result, err := calculateHMAC(data, secret, HMAC_SHA256)
	if err != nil {
		t.Errorf("HMAC_SHA256 returned an error: %v", err)
	}
	if result != expectedSHA256 {
		t.Errorf("HMAC_SHA256 result mismatch. Expected: %s, Got: %s", expectedSHA256, result)
	}

	// Test HMAC_SHA1
	result, err = calculateHMAC(data, secret, HMAC_SHA1)
	if err != nil {
		t.Errorf("HMAC_SHA1 returned an error: %v", err)
	}
	if result != expectedSHA1 {
		t.Errorf("HMAC_SHA1 result mismatch. Expected: %s, Got: %s", expectedSHA1, result)
	}

	// Test HMAC_MD5
	result, err = calculateHMAC(data, secret, HMAC_MD5)
	if err != nil {
		t.Errorf("HMAC_MD5 returned an error: %v", err)
	}
	if result != expectedMD5 {
		t.Errorf("HMAC_MD5 result mismatch. Expected: %s, Got: %s", expectedMD5, result)
	}

	// Test UNKNOWN algorithm
	result, err = calculateHMAC(data, secret, UNKNOWN)
	if err == nil {
		t.Error("UNKNOWN algorithm did not return an error")
	}
	if result != "" {
		t.Errorf("UNKNOWN algorithm returned non-empty string: %s", result)
	}
	if err.Error() != "unsupported algorithm" {
		t.Errorf("UNKNOWN algorithm returned wrong error message: %s", err.Error())
	}
}

func TestCalculateHMAC_EmptyData(t *testing.T) {
	data := ""
	secret := "test secret"

	// Compute expected values
	hmacSHA256 := hmac.New(sha256.New, []byte(secret))
	hmacSHA256.Write([]byte(data))
	expectedSHA256 := base64.StdEncoding.EncodeToString(hmacSHA256.Sum(nil))

	result, err := calculateHMAC(data, secret, HMAC_SHA256)
	if err != nil {
		t.Errorf("HMAC_SHA256 with empty data returned an error: %v", err)
	}
	if result != expectedSHA256 {
		t.Errorf("HMAC_SHA256 with empty data result mismatch. Expected: %s, Got: %s", expectedSHA256, result)
	}
}

func TestCalculateHMAC_EmptySecret(t *testing.T) {
	data := "test data"
	secret := ""

	// Compute expected values
	hmacSHA256 := hmac.New(sha256.New, []byte(secret))
	hmacSHA256.Write([]byte(data))
	expectedSHA256 := base64.StdEncoding.EncodeToString(hmacSHA256.Sum(nil))

	result, err := calculateHMAC(data, secret, HMAC_SHA256)
	if err != nil {
		t.Errorf("HMAC_SHA256 with empty secret returned an error: %v", err)
	}
	if result != expectedSHA256 {
		t.Errorf("HMAC_SHA256 with empty secret result mismatch. Expected: %s, Got: %s", expectedSHA256, result)
	}
}

func TestCalculateHMAC_SpecialCharacters(t *testing.T) {
	data := "test data with special characters: @#$%^&*()"
	secret := "secret with special characters: ñçàèö"

	// Compute expected values
	hmacSHA256 := hmac.New(sha256.New, []byte(secret))
	hmacSHA256.Write([]byte(data))
	expectedSHA256 := base64.StdEncoding.EncodeToString(hmacSHA256.Sum(nil))

	result, err := calculateHMAC(data, secret, HMAC_SHA256)
	if err != nil {
		t.Errorf("HMAC_SHA256 with special characters returned an error: %v", err)
	}
	if result != expectedSHA256 {
		t.Errorf("HMAC_SHA256 with special characters result mismatch. Expected: %s, Got: %s", expectedSHA256, result)
	}
}
