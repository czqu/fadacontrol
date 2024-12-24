package schema

import (
	"bytes"
	"encoding/binary"
	"io"
)

type InternalCommandType uint32

const (
	UnknownCommandType InternalCommandType = iota

	LockPCCommandType
	ExitCommandType
)

type InternalCommand struct {
	CommandType InternalCommandType `json:"command_type"`
	Data        interface{}         `json:"data"`
}
type InternalDataPacket struct {
	DataLength uint16
	DataType   InternalDataType
	Data       []byte
}
type InternalDataType uint16

const (
	UnknownDataType InternalDataType = iota
	JsonData
	BinaryData
)

func (p *InternalDataPacket) Pack() ([]byte, error) {

	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.BigEndian, p.DataLength)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, p.DataType)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, p.Data)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (p *InternalDataPacket) Unpack(buffer io.Reader) error {

	var length uint16
	err := binary.Read(buffer, binary.BigEndian, &length)
	if err != nil {
		return err
	}

	var dataType uint16
	err = binary.Read(buffer, binary.BigEndian, &dataType)
	if err != nil {
		return err
	}

	packetData := make([]byte, length)
	err = binary.Read(buffer, binary.BigEndian, &packetData)
	if err != nil {
		return err
	}

	p.DataLength = length
	p.Data = packetData
	p.DataType = InternalDataType(dataType)
	return nil
}
