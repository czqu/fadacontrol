package remote_schema

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fadacontrol/pkg/secure"
)

type PacketType uint8

const (
	reserve PacketType = iota
	JsonType
	ProtoBuf
	Text
)

type KeyGenAlgorithm uint8

const (
	NoSalt KeyGenAlgorithm = iota
	Arg2iD
)

type PayloadPacket struct {
	Reserve             uint8
	RequestIdLen        uint8
	RequestId           []byte
	EncryptionAlgorithm secure.EncryptionAlgorithmEnum // 1 byte encryption algorithm length combination 0x00 reserved for unencrypted 0xff
	DataType            PacketType                     // packet Type
	Data                []byte                         // data section
}

// Pack converts a PayloadPacket struct into a byte slice.
func (p *PayloadPacket) Pack() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(p.Reserve)
	// Write RequestIdLen
	if err := binary.Write(&buf, binary.BigEndian, p.RequestIdLen); err != nil {
		return nil, err
	}
	if p.RequestIdLen > 0 {
		// Write RequestId
		if _, err := buf.Write(p.RequestId); err != nil {
			return nil, err
		}
	}
	// Write EncryptionAlgorithm
	if err := binary.Write(&buf, binary.BigEndian, p.EncryptionAlgorithm); err != nil {
		return nil, err
	}

	// Write DataType
	if err := binary.Write(&buf, binary.BigEndian, p.DataType); err != nil {
		return nil, err
	}

	// Write Data
	if _, err := buf.Write(p.Data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unpack converts a byte slice into a PayloadPacket struct.
func (p *PayloadPacket) Unpack(data []byte) error {
	buf := bytes.NewReader(data)

	b, err := buf.ReadByte()
	if err != nil {
		return err
	}
	p.Reserve = b
	// Read RequestIdLen
	if err := binary.Read(buf, binary.BigEndian, &p.RequestIdLen); err != nil {
		return err
	}
	p.RequestId = nil
	if p.RequestIdLen > 0 {
		// Read RequestId
		requestId := make([]byte, p.RequestIdLen)
		if _, err := buf.Read(requestId); err != nil {
			return err
		}
		p.RequestId = requestId
	}

	// Read EncryptionAlgorithm
	if err := binary.Read(buf, binary.BigEndian, &p.EncryptionAlgorithm); err != nil {
		return err
	}

	// Read DataType
	if err := binary.Read(buf, binary.BigEndian, &p.DataType); err != nil {
		return err
	}

	// Read Data
	payload := make([]byte, buf.Len())
	if _, err := buf.Read(payload); err != nil {
		return err
	}
	p.Data = payload

	return nil
}

// PacketToBase64 converts a PayloadPacket to a base64 encoded string.
func PacketToBase64(packet *PayloadPacket) (string, error) {
	// Pack the packet into a byte slice
	data, err := packet.Pack()
	if err != nil {
		return "", err
	}

	// Encode the byte slice to base64
	base64Str := base64.StdEncoding.EncodeToString(data)
	return base64Str, nil
}

// Base64ToPacket converts a base64 encoded string to a PayloadPacket.
func Base64ToPacket(base64Str string) (*PayloadPacket, error) {
	// Decode the base64 string to a byte slice
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}
	packet := &PayloadPacket{}
	// Unpack the byte slice into a PayloadPacket
	err = packet.Unpack(data)
	if err != nil {
		return nil, err
	}

	return packet, nil
}
