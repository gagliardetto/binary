package bin

import "encoding/binary"

type option struct {
	OptionalField bool
	SizeOfSlice   *int
	Order         binary.ByteOrder
}

func LE() binary.ByteOrder { return binary.LittleEndian }
func BE() binary.ByteOrder { return binary.BigEndian }

func newDefaultOption() *option {
	return &option{
		OptionalField: false,
		Order:         LE(),
	}
}

func (o *option) clone() *option {
	out := &option{
		OptionalField: o.OptionalField,
		SizeOfSlice:   o.SizeOfSlice,
		Order:         o.Order,
	}
	return out
}
func (o *option) isOptional() bool {
	return o.OptionalField
}

func (o *option) hasSizeOfSlice() bool {
	return o.SizeOfSlice != nil
}

func (o *option) getSizeOfSlice() int {
	return *o.SizeOfSlice
}

func (o *option) setSizeOfSlice(size int) *option {
	o.SizeOfSlice = &size
	return o
}
func (o *option) setIsOptional(isOptional bool) *option {
	o.OptionalField = isOptional
	return o
}

type Encoding int

const (
	EncodingBin Encoding = iota
	EncodingCompactU16
	EncodingBorsh
)

func (enc Encoding) String() string {
	switch enc {
	case EncodingBin:
		return "Bin"
	case EncodingCompactU16:
		return "CompactU16"
	case EncodingBorsh:
		return "Borsh"
	default:
		return ""
	}
}

func (en Encoding) IsBorsh() bool {
	return en == EncodingBorsh
}

func (en Encoding) IsBin() bool {
	return en == EncodingBin
}

func (en Encoding) IsCompactU16() bool {
	return en == EncodingCompactU16
}

func isValidEncoding(enc Encoding) bool {
	switch enc {
	case EncodingBin, EncodingCompactU16, EncodingBorsh:
		return true
	default:
		return false
	}
}
