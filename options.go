package bin



type Option struct {
	optionalField bool
	sizeOfSlice   *int
}

func (o *Option) isOptional() bool {
	return o.optionalField
}

func (o *Option) hasSizeOfSlice() bool {
	return o.sizeOfSlice != nil
}

func (o *Option) getSizeOfSlice() int {
	return *o.sizeOfSlice
}

func (o *Option) setSizeOfSlice(size int) {
	o.sizeOfSlice = &size
}

