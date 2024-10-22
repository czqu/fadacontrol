package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type PipePacket struct {
	Tpe  PipePacketType
	Size uint32
	Data []byte
}

type PipePacketType uint8

const (
	Unknown PipePacketType = iota
	Resp
	UnlockReq
	SetFieldBitmap
	CommandClicked
	SetCommandClickText
	SetLargeText
	END = 0xff
)

// PipePacket 的 pack 方法，将其序列化为字节流
func (p *PipePacket) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 写入类型（Tpe） - 1字节
	buf.WriteByte(byte(p.Tpe))

	// 写入Size（Size） - 4字节
	err := binary.Write(buf, binary.BigEndian, p.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to write Size: %v", err)
	}

	// 写入Data
	buf.Write(p.Data)

	return buf.Bytes(), nil
}

func (p *PipePacket) Unpack(buf io.Reader) error {

	// 读取类型（Tpe） - 1字节
	err := binary.Read(buf, binary.BigEndian, &p.Tpe)
	if err != nil {
		return err
	}

	// 读取Size（Size） - 4字节
	err = binary.Read(buf, binary.BigEndian, &p.Size)
	if err != nil {
		return fmt.Errorf("failed to read Size: %v", err)
	}

	// 读取Data，根据Size读取固定字节
	p.Data = make([]byte, p.Size)
	_, err = buf.Read(p.Data)
	if err != nil {
		return err
	}

	return nil
}
