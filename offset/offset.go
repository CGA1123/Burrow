package offset

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Offset contains the result of a single offset message, representing an offset
// commit for a particular consumer group topic partition.
type Offset struct {
	ConsumerGroup string
	Topic         string
	Partition     int32
	Offset        int64
	Timestamp     int64
}

type offsetKey struct {
	ConsumerGroup string
	Topic         string
	Partition     int32
}

// ErrGroupMetadataMsg is returned by Decode when a metadata message is
// attempted decoding (key version = 2). Metadata decoding is not implemented.
var ErrGroupMetadataMsg = fmt.Errorf("group metadata message")

// Decode parses a consumer offset message
func Decode(key, value []byte) (*Offset, error) {
	keybuf, valbuf := bytes.NewBuffer(key), bytes.NewBuffer(value)
	keyVersion, err := readInt16(keybuf)
	if err != nil {
		return nil, fmt.Errorf("error reader key version: %v", err)
	}

	switch keyVersion {
	case 0, 1:
		return DecodeOffset(keybuf, valbuf)
	case 2:
		return nil, ErrGroupMetadataMsg
	default:
		return nil, fmt.Errorf("unknown key version: %v", keyVersion)
	}
}

func DecodeOffset(keybuf, valbuf *bytes.Buffer) (*Offset, error) {
	key, err := decodeKey(keybuf)
	if err != nil {
		return nil, fmt.Errorf("error decoding offset key: %v", err)
	}

	valueVersion, err := readInt16(valbuf)
	if err != nil {
		return nil, fmt.Errorf("error decoding value version: %v", err)
	}

	var decoder func(*bytes.Buffer) (int64, int64, error)
	switch valueVersion {
	case 0, 1:
		decoder = offsetValueDecoderV0
	case 3:
		decoder = offsetValueDecoderV3
	default:
		return nil, fmt.Errorf("unknown value version: %v", valueVersion)
	}

	offset, timestamp, err := decoder(valbuf)
	if err != nil {
		return nil, fmt.Errorf("error decoding offset value: %v", err)
	}

	return &Offset{
		ConsumerGroup: key.ConsumerGroup,
		Topic:         key.Topic,
		Partition:     key.Partition,
		Offset:        offset,
		Timestamp:     timestamp,
	}, nil
}

func offsetValueDecoderV0(valbuf *bytes.Buffer) (int64, int64, error) {
	offset, err := readInt64(valbuf)
	if err != nil {
		return 0, 0, fmt.Errorf("error decoding offset: %v", err)
	}

	if _, err := readString(valbuf); err != nil {
		return 0, 0, fmt.Errorf("error decoding metadata: %v", err)
	}

	timestamp, err := readInt64(valbuf)
	if err != nil {
		return 0, 0, fmt.Errorf("error decoding timestamp: %v", err)
	}

	return offset, timestamp, nil
}

func offsetValueDecoderV3(valbuf *bytes.Buffer) (int64, int64, error) {
	offset, err := readInt64(valbuf)
	if err != nil {
		return 0, 0, fmt.Errorf("error decoding offset: %v", err)
	}

	if _, err := readInt32(valbuf); err != nil {
		return 0, 0, fmt.Errorf("error decoding leader epoch: %v", err)
	}

	if _, err := readString(valbuf); err != nil {
		return 0, 0, fmt.Errorf("error decoding metadata: %v", err)
	}

	timestamp, err := readInt64(valbuf)
	if err != nil {
		return 0, 0, fmt.Errorf("error decoding timestamp: %v", err)
	}

	return offset, timestamp, nil
}

func decodeKey(keybuf *bytes.Buffer) (*offsetKey, error) {
	key := &offsetKey{}

	var err error

	key.ConsumerGroup, err = readString(keybuf)
	if err != nil {
		return nil, fmt.Errorf("error decoding consumer group: %v", err)
	}

	key.Topic, err = readString(keybuf)
	if err != nil {
		return nil, fmt.Errorf("error decoding topic: %v", err)
	}

	key.Partition, err = readInt32(keybuf)
	if err != nil {
		return nil, fmt.Errorf("error decoding partition: %v", err)
	}

	return key, nil
}

// Kafka protocol format readers
// See: https://bit.ly/359d3Yd

func readInt8(buf *bytes.Buffer) (int8, error) {
	var value int8
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return value, fmt.Errorf("error reading int8: %v", err)
	}

	return value, nil
}

func readInt16(buf *bytes.Buffer) (int16, error) {
	var value int16
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return value, fmt.Errorf("error reading int16: %v", err)
	}

	return value, nil
}

func readInt32(buf *bytes.Buffer) (int32, error) {
	var value int32
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return value, fmt.Errorf("error reading int32: %v", err)
	}

	return value, nil
}

func readInt64(buf *bytes.Buffer) (int64, error) {
	var value int64
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return value, fmt.Errorf("error reading int64: %v", err)
	}

	return value, nil
}

func readString(buf *bytes.Buffer) (string, error) {
	length, err := readInt16(buf)
	if err != nil {
		return "", fmt.Errorf("error reading string header: %v", err)
	}

	if length == -1 {
		return "", nil
	}

	data := buf.Next(int(length))
	if len(data) != int(length) {
		return "", fmt.Errorf("readString underflow: expected %v got %v", length, len(data))
	}

	return string(data), nil
}

func readBytes(buf *bytes.Buffer) ([]byte, error) {
	length, err := readInt32(buf)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes header: %v", err)
	}

	if length == -1 {
		return nil, nil
	}

	data := buf.Next(int(length))
	if len(data) != int(length) {
		return nil, fmt.Errorf("readBytes underflow: expected %v got %v", length, len(data))
	}

	cp := make([]byte, length)
	copy(cp, data)

	return cp, nil
}
