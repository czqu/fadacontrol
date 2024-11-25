package remote_schema

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPayloadPacket_PackUnpack(t *testing.T) {
	originalPacket := &PayloadPacket{
		SessionIdLen: 4,
		SessionId:    []byte("test"),
		DataType:     JsonType,
		Data:         []byte("sample data"),
	}

	// Test Pack
	packedData, err := originalPacket.Pack()
	assert.NoError(t, err, "Packing should not produce an error")

	// Test Unpack
	unpackedPacket := &PayloadPacket{}
	err = unpackedPacket.Unpack(packedData)
	assert.NoError(t, err, "Unpacking should not produce an error")

	// Verify the unpacked data matches the original
	assert.Equal(t, originalPacket.SessionIdLen, unpackedPacket.SessionIdLen, "SessionIdLen should match")
	assert.Equal(t, originalPacket.SessionId, unpackedPacket.SessionId, "SessionId should match")
	assert.Equal(t, originalPacket.DataType, unpackedPacket.DataType, "DataType should match")
	assert.Equal(t, originalPacket.Data, unpackedPacket.Data, "Data should match")
}

func TestPayloadPacket_UnpackError(t *testing.T) {
	packet := &PayloadPacket{}

	// Simulate an error in the Unpack method by providing incomplete data
	incompleteData := []byte{0x01, 0x02} // Not enough data
	err := packet.Unpack(incompleteData)
	assert.Error(t, err, "Unpacking should produce an error with incomplete data")
}
func TestBase64ToPacket(t *testing.T) {
	// Test Base64ToPacket
	originalPacket := &PayloadPacket{
		SessionIdLen: 4,
		SessionId:    []byte("test"),
		DataType:     JsonType,
		Data:         []byte("sample data"),
	}
	packedData, err := originalPacket.Pack()
	assert.NoError(t, err, "Packing should not produce an error")
	packedDataBase64Std := base64.StdEncoding.EncodeToString(packedData)
	packedDataBase64, err := PacketToBase64(originalPacket)
	assert.NoError(t, err, "Unpacking should not produce an error")
	assert.Equal(t, packedDataBase64Std, packedDataBase64, "Base64 payload packet does not match")

	packet, err := Base64ToPacket(packedDataBase64Std)
	assert.NoError(t, err, "Unpacking should not produce an error")
	assert.Equal(t, originalPacket.SessionIdLen, packet.SessionIdLen, "SessionIdLen should match")
	assert.Equal(t, originalPacket.SessionId, packet.SessionId, "SessionId should match")
	assert.Equal(t, originalPacket.DataType, packet.DataType, "DataType should match")
	assert.Equal(t, originalPacket.Data, packet.Data, "Data should match")

}
func TestPayloadPacket_PackUnpack_NoRequestId(t *testing.T) {
	originalPacket := &PayloadPacket{
		SessionIdLen: 0,
		SessionId:    nil,
		DataType:     JsonType,
		Data:         []byte("sample data"),
	}

	// Test Pack
	packedData, err := originalPacket.Pack()
	assert.NoError(t, err, "Packing should not produce an error")

	// Test Unpack
	unpackedPacket := &PayloadPacket{}
	err = unpackedPacket.Unpack(packedData)
	assert.NoError(t, err, "Unpacking should not produce an error")

	// Verify the unpacked data matches the original
	assert.Equal(t, originalPacket.SessionIdLen, unpackedPacket.SessionIdLen, "SessionIdLen should match")
	assert.Equal(t, originalPacket.SessionId, unpackedPacket.SessionId, "SessionId should match")
	assert.Equal(t, originalPacket.DataType, unpackedPacket.DataType, "DataType should match")
	assert.Equal(t, originalPacket.Data, unpackedPacket.Data, "Data should match")
}
