package bin

import "bytes"

type MarshalerBinary interface {
	MarshalBinary(encoder *Encoder) error
}

type UnmarshalerBinary interface {
	UnmarshalBinary(decoder *Decoder) error
}

func MarshalBinary(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := NewEncoder(buf)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}
