package bin

import (
	"bytes"
	"fmt"
)

type MarshalerBinary interface {
	MarshalBinary(encoder *Encoder) error
}

type UnmarshalerBinary interface {
	UnmarshalBinary(decoder *Decoder) error
}

func MarshalBin(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := NewBinEncoder(buf)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}

func MarshalBorsh(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := NewBorshEncoder(buf)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}

func MarshalCompact16(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := NewCompact16Encoder(buf)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}

func UnmarshalBin(v interface{}, b []byte) error {
	decoder := NewBinDecoder(b)
	return decoder.Decode(v)
}

func UnmarshalBorsh(v interface{}, b []byte) error {
	decoder := NewBorshDecoder(b)
	return decoder.Decode(v)
}

func UnmarshalCompact16(v interface{}, b []byte) error {
	decoder := NewCompact16Decoder(b)
	return decoder.Decode(v)
}

type byteCounter struct {
	count uint64
}

func (c *byteCounter) Write(p []byte) (n int, err error) {
	c.count += uint64(len(p))
	return len(p), nil
}

// BinByteCount computes the byte count size for the received populated structure. The reported size
// is the one for the populated structure received in arguments. Depending on how serialization of
// your fields is performed, size could vary for different structure.
func BinByteCount(v interface{}) (uint64, error) {
	counter := byteCounter{}
	err := NewBinEncoder(&counter).Encode(v)
	if err != nil {
		return 0, fmt.Errorf("encode %T: %w", v, err)
	}
	return counter.count, nil
}

// BorshByteCount computes the byte count size for the received populated structure. The reported size
// is the one for the populated structure received in arguments. Depending on how serialization of
// your fields is performed, size could vary for different structure.
func BorshByteCount(v interface{}) (uint64, error) {
	counter := byteCounter{}
	err := NewBorshEncoder(&counter).Encode(v)
	if err != nil {
		return 0, fmt.Errorf("encode %T: %w", v, err)
	}
	return counter.count, nil
}

// Compact16ByteCount computes the byte count size for the received populated structure. The reported size
// is the one for the populated structure received in arguments. Depending on how serialization of
// your fields is performed, size could vary for different structure.
func Compact16ByteCount(v interface{}) (uint64, error) {
	counter := byteCounter{}
	err := NewCompact16Encoder(&counter).Encode(v)
	if err != nil {
		return 0, fmt.Errorf("encode %T: %w", v, err)
	}
	return counter.count, nil
}

// MustBinByteCount acts just like BinByteCount but panics if it encounters any encoding errors.
func MustBinByteCount(v interface{}) uint64 {
	count, err := BinByteCount(v)
	if err != nil {
		panic(err)
	}
	return count
}

// MustBorshByteCount acts just like BorshByteCount but panics if it encounters any encoding errors.
func MustBorshByteCount(v interface{}) uint64 {
	count, err := BorshByteCount(v)
	if err != nil {
		panic(err)
	}
	return count
}

// MustCompact16ByteCount acts just like Compact16ByteCount but panics if it encounters any encoding errors.
func MustCompact16ByteCount(v interface{}) uint64 {
	count, err := Compact16ByteCount(v)
	if err != nil {
		panic(err)
	}
	return count
}
