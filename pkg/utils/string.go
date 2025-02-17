package utils

import (
	"errors"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenRandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func ConvertToString(r interface{}) string {
	switch v := r.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return "unknown panic"
	}
}
func ConvertToError(r interface{}) error {
	e, ok := r.(error)
	if !ok {
		e = errors.New(ConvertToString(r))
	}
	return e
}
