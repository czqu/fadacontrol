package remote_schema

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fadacontrol/pkg/secure"
	"fmt"
	"golang.org/x/crypto/argon2"
)

type PacketType uint8

const (
	reserve PacketType = iota
	JsonType
)

type EncryptionAlgorithmEnum uint8

const MaxKeyLength = 32
const (
	None               EncryptionAlgorithmEnum = iota
	AESGCM128Algorithm                         // The AES-128GCM key is 16 bytes long
	AESGCM192Algorithm                         // The AES-192GCM key is 24 bytes long
	AESGCM256Algorithm                         // The AES-256GCM key is 32 bytes long
	Unknown            = 0xff
)

type KeyGenAlgorithm uint8

const (
	NoSalt KeyGenAlgorithm = iota
	Arg2iD
)

var AlgorithmKeyLengths = map[EncryptionAlgorithmEnum]int{
	None:               0,  // No encryption, no key length
	AESGCM128Algorithm: 16, // 128-bit AES-GCM key length
	AESGCM192Algorithm: 24, // 192-bit AES-GCM key length
	AESGCM256Algorithm: 32, // 256-bit AES-GCM key length
	Unknown:            -1, // Unknown encryption algorithm
}

type PayloadPacket struct {
	EncryptionAlgorithm EncryptionAlgorithmEnum // 1字节 加密算法长度组合 0x00为不加密 0xff 保留
	KeyGenAlgorithm     KeyGenAlgorithm         // 1 字节 密钥生成算法 0为直接使用密码作为密钥
	SaltLength          uint16                  // 2字节 密钥生成函数所需的盐长度 单位字节 可为0 为零代表客户端直接已知密钥或者盐，将使用系统自带盐
	Salt                []byte                  //  盐
	DataType            PacketType              //数据包类型:
	Data                []byte                  // 数据部分
}

// Pack converts a PayloadPacket struct into a byte slice.
func (p *PayloadPacket) Pack() ([]byte, error) {
	var buf bytes.Buffer

	// Write EncryptionAlgorithm
	if err := binary.Write(&buf, binary.BigEndian, p.EncryptionAlgorithm); err != nil {
		return nil, err
	}

	// Write KeyGenAlgorithm
	if err := binary.Write(&buf, binary.BigEndian, p.KeyGenAlgorithm); err != nil {
		return nil, err
	}

	// Write SaltLength
	if err := binary.Write(&buf, binary.BigEndian, p.SaltLength); err != nil {
		return nil, err
	}

	// Write Salt
	if _, err := buf.Write(p.Salt); err != nil {
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

	// Read EncryptionAlgorithm
	if err := binary.Read(buf, binary.BigEndian, &p.EncryptionAlgorithm); err != nil {
		return err
	}

	// Read KeyGenAlgorithm
	if err := binary.Read(buf, binary.BigEndian, &p.KeyGenAlgorithm); err != nil {
		return err
	}

	// Read SaltLength
	if err := binary.Read(buf, binary.BigEndian, &p.SaltLength); err != nil {
		return err
	}

	// Read Salt
	p.Salt = make([]byte, p.SaltLength)
	if _, err := buf.Read(p.Salt); err != nil {
		return err
	}

	// Read DataType
	if err := binary.Read(buf, binary.BigEndian, &p.DataType); err != nil {
		return err
	}

	// Read Data
	p.Data = make([]byte, buf.Len())
	if _, err := buf.Read(p.Data); err != nil {
		return err
	}

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
func EncryptData(inputSecret string, salt []byte, data []byte) ([]byte, error) {
	key256, err := DeriveKey(inputSecret, salt, AlgorithmKeyLengths[AESGCM128Algorithm])
	if err != nil {
		return nil, fmt.Errorf("error deriving key: %v", err)
	}

	encryptedData, err := secure.EncryptAESGCM(key256, data)
	if err != nil {
		return nil, fmt.Errorf("encryption error: %v", err)
	}

	packet := PayloadPacket{
		EncryptionAlgorithm: AESGCM128Algorithm,
		KeyGenAlgorithm:     Arg2iD,
		SaltLength:          uint16(len(salt)),
		Salt:                salt,
		DataType:            JsonType,
		Data:                encryptedData,
	}
	return packet.Pack()
}
func DecryptData(encryptData, salt []byte, inputSecret string) ([]byte, error) {

	// Derive the key using the same method as in the encryption function
	key, err := DeriveKey(inputSecret, salt, AlgorithmKeyLengths[AESGCM128Algorithm])
	if err != nil {
		return nil, fmt.Errorf("error deriving key: %v", err)
	}

	// Decrypt the data using the key and the encrypted data from the packet
	decryptedData, err := secure.DecryptAESGCM(key, encryptData)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}

	return decryptedData, nil
}

// DeriveKey derives an AES key of the specified size from a password using Argon2 KDF.
func DeriveKey(password string, salt []byte, keySize int) ([]byte, error) {

	key := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, uint32(keySize))
	return key, nil
}
