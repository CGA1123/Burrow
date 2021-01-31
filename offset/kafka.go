package offset

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Kafka protocol format readers
// See: https://bit.ly/359d3Yd

func ReadInt8(buf io.Reader) (int8, error) {
	var value int8
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return value, fmt.Errorf("error reading int8: %v", err)
	}

	return value, nil
}

func ReadInt16(buf io.Reader) (int16, error) {
	var value int16
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return value, fmt.Errorf("error reading int16: %v", err)
	}

	return value, nil
}

func ReadInt32(buf io.Reader) (int32, error) {
	var value int32
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return value, fmt.Errorf("error reading int32: %v", err)
	}

	return value, nil
}

func ReadInt64(buf io.Reader) (int64, error) {
	var value int64
	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return value, fmt.Errorf("error reading int64: %v", err)
	}

	return value, nil
}

func ReadString(buf *bytes.Buffer) (string, error) {
	length, err := ReadInt16(buf)
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

func ReadBytes(buf *bytes.Buffer) ([]byte, error) {
	length, err := ReadInt32(buf)
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
