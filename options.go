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

func (o *option) isOptional() bool {
	return o.OptionalField
}

func (o *option) hasSizeOfSlice() bool {
	return o.SizeOfSlice != nil
}

func (o *option) getSizeOfSlice() int {
	return *o.SizeOfSlice
}

func (o *option) setSizeOfSlice(size int) {
	o.SizeOfSlice = &size
}

type Encoding int

const (
	EncodingBin Encoding = iota
	EncodingCompact16
	EncodingBorsh
)

var Encodings = struct {
	Bin       Encoding
	Compact16 Encoding
	Borsh     Encoding
}{
	Bin:       EncodingBin,
	Compact16: EncodingCompact16,
	Borsh:     EncodingBorsh,
}

func (enc Encoding) String() string {
	switch enc {
	case Encodings.Bin:
		return "Bin"
	case Encodings.Compact16:
		return "Compact16"
	case Encodings.Borsh:
		return "Borsh"
	default:
		return ""
	}
}

func (en Encoding) IsBorsh() bool {
	return en == Encodings.Borsh
}

func (en Encoding) IsBin() bool {
	return en == Encodings.Bin
}

func (en Encoding) IsCompact16() bool {
	return en == Encodings.Compact16
}

func isValidEncoding(enc Encoding) bool {
	switch enc {
	case Encodings.Bin, Encodings.Compact16, Encodings.Borsh:
		return true
	default:
		return false
	}
}
