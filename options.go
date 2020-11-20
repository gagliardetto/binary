package bin

import "encoding/binary"

type option struct {
	OptionalField bool
	SizeOfSlice   *int
	Order         binary.ByteOrder
}

func LE() *option { return &option{Order: binary.LittleEndian} }
func BE() *option { return &option{Order: binary.BigEndian} }

func newDefaultOption() *option {
	return &option{
		OptionalField: false,
		Order:         binary.LittleEndian,
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
