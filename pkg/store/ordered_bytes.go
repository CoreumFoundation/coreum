package store

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

const (
	// Uint8OrderedBytesSize is byte size required to store ordered uint8.
	Uint8OrderedBytesSize = 1
	// Int8OrderedBytesSize is byte size required to store ordered int8.
	Int8OrderedBytesSize = Uint8OrderedBytesSize
	// Uint64OrderedBytesSize is byte size required to store ordered uint64.
	Uint64OrderedBytesSize = 8
)

// AppendInt8ToOrderedBytes appends int8 ordered bytes to bytes array.
func AppendInt8ToOrderedBytes(b []byte, v int8) []byte {
	// int8 -> -128 to 127
	// uint8 -> 0 to 255
	uint8v := uint8(int16(v) + int16(128))
	return AppendUint8ToOrderedBytes(b, uint8v)
}

// ReadOrderedBytesToInt8 returns ordered bytes array converted into int8 and remaining bytes.
func ReadOrderedBytesToInt8(b []byte) (int8, []byte, error) {
	uint8v, remB, err := ReadOrderedBytesToUint8(b)
	if err != nil {
		return 0, nil, err
	}
	int8v := int8(int16(uint8v) - int16(128))
	return int8v, remB, nil
}

// AppendUint8ToOrderedBytes converts uint8 into ordered bytes and appends it to bytes array.
func AppendUint8ToOrderedBytes(b []byte, v uint8) []byte {
	return append(b, v)
}

// ReadOrderedBytesToUint8 returns ordered bytes array converted into uint8 and remaining bytes.
func ReadOrderedBytesToUint8(b []byte) (uint8, []byte, error) {
	if len(b) < Uint8OrderedBytesSize {
		return 0, nil, errors.Errorf("invalid bytes length, min %d", Uint8OrderedBytesSize)
	}
	return b[0], b[Uint8OrderedBytesSize:], nil
}

// AppendUint64ToOrderedBytes converts uint64 into ordered bytes and appends it to bytes array.
func AppendUint64ToOrderedBytes(b []byte, v uint64) []byte {
	return binary.BigEndian.AppendUint64(b, v)
}

// ReadOrderedBytesToUint64 returns ordered bytes array converted into uint64 and remaining bytes.
func ReadOrderedBytesToUint64(b []byte) (uint64, []byte, error) {
	if len(b) < Uint64OrderedBytesSize {
		return 0, nil, errors.Errorf("invalid bytes length, min %d", Uint64OrderedBytesSize)
	}
	return binary.BigEndian.Uint64(b[:Uint64OrderedBytesSize]), b[Uint64OrderedBytesSize:], nil
}
