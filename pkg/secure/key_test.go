package secure

import (
	"testing"
)

func TestGenerateRandomBase58Key(t *testing.T) {
	_, err := GenerateRandomBase58Key(35)
	if err != nil {
		t.Error(err.Error())
	}

}
