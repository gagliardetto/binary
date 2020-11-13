package bin

type Option struct {
	OptionalField bool
	SizeOfSlice   *int
}

func (o *Option) isOptional() bool {
	return o.OptionalField
}

func (o *Option) hasSizeOfSlice() bool {
	return o.SizeOfSlice != nil
}

func (o *Option) getSizeOfSlice() int {
	return *o.SizeOfSlice
}

func (o *Option) setSizeOfSlice(size int) {
	o.SizeOfSlice = &size
}
