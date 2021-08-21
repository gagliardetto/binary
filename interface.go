package bin

import (
	"bytes"
	"fmt"
)

type BinaryMarshaler interface {
	MarshalWithEncoder(encoder *Encoder) error
}

type BinaryUnmarshaler interface {
	UnmarshalWithDecoder(decoder *Decoder) error
}

type EncoderDecoder interface {
	BinaryMarshaler
	BinaryUnmarshaler
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

func MarshalCompactU16(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := NewCompactU16Encoder(buf)
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

func UnmarshalCompactU16(v interface{}, b []byte) error {
	decoder := NewCompactU16Decoder(b)
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

// CompactU16ByteCount computes the byte count size for the received populated structure. The reported size
// is the one for the populated structure received in arguments. Depending on how serialization of
// your fields is performed, size could vary for different structure.
func CompactU16ByteCount(v interface{}) (uint64, error) {
	counter := byteCounter{}
	err := NewCompactU16Encoder(&counter).Encode(v)
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

// MustCompactU16ByteCount acts just like CompactU16ByteCount but panics if it encounters any encoding errors.
func MustCompactU16ByteCount(v interface{}) uint64 {
	count, err := CompactU16ByteCount(v)
	if err != nil {
		panic(err)
	}
	return count
}
