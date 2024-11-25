package remote_schema

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
)

type PacketType uint8

const (
	reserve PacketType = iota
	JsonType
	ProtoBuf
	Text
)

type PayloadPacket struct {
	SessionIdLen uint8
	SessionId    []byte
	DataType     PacketType // packet Type
	Data         []byte     // data section
}

// Pack converts a PayloadPacket struct into a byte slice.
func (p *PayloadPacket) Pack() ([]byte, error) {
	var buf bytes.Buffer

	// Write SessionIdLen
	if err := binary.Write(&buf, binary.BigEndian, p.SessionIdLen); err != nil {
		return nil, err
	}
	if p.SessionIdLen > 0 {
		// Write SessionId
		if _, err := buf.Write(p.SessionId); err != nil {
			return nil, err
		}
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

	// Read SessionIdLen
	if err := binary.Read(buf, binary.BigEndian, &p.SessionIdLen); err != nil {
		return err
	}
	p.SessionId = nil
	if p.SessionIdLen > 0 {
		// Read SessionId
		topic := make([]byte, p.SessionIdLen)
		if _, err := buf.Read(topic); err != nil {
			return err
		}
		p.SessionId = topic
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
