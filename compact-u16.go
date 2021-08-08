package bin

import "io"

// EncodeLength encodes a "Compact-u16" length into the provided slice pointer.
// See https://docs.solana.com/developing/programming-model/transactions#compact-u16-format
// See https://github.com/solana-labs/solana/blob/2ef2b6daa05a7cff057e9d3ef95134cee3e4045d/web3.js/src/util/shortvec-encoding.ts
func EncodeLength(bytes *[]byte, ln int) {
	rem_len := ln
	for {
		elem := rem_len & 0x7f
		rem_len >>= 7
		if rem_len == 0 {
			*bytes = append(*bytes, byte(elem))
			break
		} else {
			elem |= 0x80
			*bytes = append(*bytes, byte(elem))
		}
	}
}

// DecodeLength decodes a "Compact-u16" length from the provided byte slice.
func DecodeLength(bytes []byte) int {
	ln := 0
	size := 0
	for {
		elem := int(bytes[0])
		bytes = bytes[1:]
		ln |= (elem & 0x7f) << (size * 7)
		size += 1
		if (elem & 0x80) == 0 {
			break
		}
	}
	return ln
}

// DecodeLengthFromByteReader decodes a "Compact-u16" length from the provided io.ByteReader.
func DecodeLengthFromByteReader(reader io.ByteReader) (int, error) {
	ln := 0
	size := 0
	for {
		elemByte, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}
		elem := int(elemByte)
		ln |= (elem & 0x7f) << (size * 7)
		size += 1
		if (elem & 0x80) == 0 {
			break
		}
	}
	return ln, nil
}
