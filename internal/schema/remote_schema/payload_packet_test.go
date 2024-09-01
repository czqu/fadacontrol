package remote_schema

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	p, err := Base64ToPacket("AQEAEAAAAAAAAAAAAAAAAAAAAAABZy5ag0cISZVzOQc+68ZId1eJFY6i+7o/nI1uZpbmdPX/O6/se0sci8ZKFUm1yhKH6Mo=")
	p.Pack()
	if err != nil {
		t.Error(err)
	}
	data, err := DecryptData(p.Data, p.Salt, "Tw01ZD22")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(data))
}
