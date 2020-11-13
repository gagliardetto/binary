package bin

import "bytes"

// MarshalerBinary is the interface implemented by types
// that can marshal to an EOSIO binary description of themselves.
//
// **Warning** This is experimental, exposed only for internal usage for now.
type MarshalerBinary interface {
	MarshalBinary(encoder *Encoder) error
}

func MarshalBinary(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := NewEncoder(buf)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}

