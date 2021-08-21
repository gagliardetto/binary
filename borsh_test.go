package bin

import (
	"bytes"
	"math"
	"reflect"
	strings2 "strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type Struct struct {
	Foo string
	Bar uint32
}

type AA struct {
	A        int64
	B        int32
	C        bool
	D        *bool
	E        *uint64
	M        map[string]string
	EmptyMap map[int64]string
	Array    [2]int64

	Optional *Struct
	Value    Struct

	InterfaceEncoderDecoderByValue   CustomEncoding
	InterfaceEncoderDecoderByPointer *CustomEncoding

	InterfaceEncoderDecoderByValueEmpty   CustomEncoding
	InterfaceEncoderDecoderByPointerEmpty *CustomEncoding

	HighValuesInt64   []int64
	HighValuesUint64  []uint64
	HighValuesFloat64 []float64
}

type CustomEncoding struct {
	Prefix byte
	Value  uint32
}

func (e *CustomEncoding) MarshalBinary(encoder *Encoder) error {
	if err := encoder.WriteByte(e.Prefix); err != nil {
		return err
	}
	return encoder.WriteUint32(e.Value, LE())
}

func (e *CustomEncoding) UnmarshalBinary(decoder *Decoder) (err error) {
	if e.Prefix, err = decoder.ReadByte(); err != nil {
		return err
	}
	if e.Value, err = decoder.ReadUint32(LE()); err != nil {
		return err
	}
	return nil
}
func TestBorsh_kitchenSink(t *testing.T) {
	boolTrue := true
	uint64Num := uint64(25464132585)
	x := AA{
		A:     1,
		B:     32,
		C:     true,
		D:     &boolTrue,
		E:     &uint64Num,
		M:     map[string]string{"foo": "bar"},
		Array: [2]int64{57, 88},
		Optional: &Struct{
			Foo: "optional foo",
			Bar: 8888886,
		},
		Value: Struct{
			Foo: "value foo",
			Bar: 7777,
		},
		InterfaceEncoderDecoderByValue:   CustomEncoding{Prefix: byte('b'), Value: 72},
		InterfaceEncoderDecoderByPointer: &CustomEncoding{Prefix: byte('c'), Value: 9999},

		HighValuesInt64: []int64{
			math.MaxInt8,
			math.MaxInt16,
			math.MaxInt32,
			math.MaxInt64,

			-math.MaxInt8,
			-math.MaxInt16,
			-math.MaxInt32,
			-math.MaxInt64,

			math.MaxUint8,
			math.MaxUint16,
			math.MaxUint32,
			// math.MaxUint64,

			-math.MaxUint8,
			-math.MaxUint16,
			-math.MaxUint32,
			// -math.MaxUint64,
		},

		HighValuesUint64: []uint64{
			math.MaxInt8,
			math.MaxInt16,
			math.MaxInt32,
			math.MaxInt64,

			math.MaxUint8,
			math.MaxUint16,
			math.MaxUint32,
			math.MaxUint64,
		},

		HighValuesFloat64: []float64{
			math.MaxFloat32,
			math.MaxFloat64,

			-math.MaxFloat32,
			-math.MaxFloat64,
		},
	}
	borshBuf := new(bytes.Buffer)
	borshEnc := NewBorshEncoder(borshBuf)
	err := borshEnc.Encode(x)
	require.NoError(t, err)

	y := new(AA)
	err = UnmarshalBorsh(y, borshBuf.Bytes())
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type A struct {
	A int64
	B int32
	C bool
	D *bool
	E *uint64
}

func TestSimple(t *testing.T) {
	boolTrue := true
	uint64Num := uint64(25464132585)
	x := A{
		A: 1,
		B: 32,
		C: true,
		D: &boolTrue,
		E: &uint64Num,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)
	y := new(A)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type B struct {
	I8         int8
	I16        int16
	I32        int32
	I64        int64
	U8         uint8
	U16        uint16
	U32        uint32
	U64        uint64
	F32        float32
	F64        float64
	unexported int64 // unexported fields are skipped.
	Err        error // nil interfaces must be specified to be skipped.
}

func TestBasic(t *testing.T) {
	x := B{
		I8:         12,
		I16:        -1,
		I32:        124,
		I64:        1243,
		U8:         1,
		U16:        979,
		U32:        123124,
		U64:        1135351135,
		F32:        -231.23,
		F64:        3121221.232,
		unexported: 333,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)
	y := new(B)

	// expect the unexported field to be zero because
	// it shouldn't have been encoded or be tried to be decoded:
	x.unexported = 0

	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type C struct {
	A3 [3]int
	S  []int
	P  *int
	M  map[string]string
}

func TestBasicContainer(t *testing.T) {
	ip := new(int)
	*ip = 213
	x := C{
		A3: [3]int{234, -123, 123},
		S:  []int{21442, 421241241, 2424},
		P:  ip,
		M:  map[string]string{"foo": "bar"},
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(C)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type N struct {
	B B
	C C
}

func TestNested(t *testing.T) {
	ip := new(int)
	*ip = 213
	x := N{
		B: B{
			I8:  12,
			I16: -1,
			I32: 124,
			I64: 1243,
			U8:  1,
			U16: 979,
			U32: 123124,
			U64: 1135351135,
			F32: -231.23,
			F64: 3121221.232,
		},
		C: C{
			A3: [3]int{234, -123, 123},
			S:  []int{21442, 421241241, 2424},
			P:  ip,
			M:  map[string]string{"foo": "bar"},
		},
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(N)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type Dummy BorshEnum

const (
	x Dummy = iota
	y
	z
)

type D struct {
	D Dummy
}

func TestSimpleEnum(t *testing.T) {
	x := D{
		D: y,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(D)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)

	require.Equal(t, x, *y)
}

type ComplexEnum struct {
	Enum BorshEnum `borsh_enum:"true"`
	Foo  Foo
	Bar  Bar
}

type Foo struct {
	FooA int32
	FooB string
}

type Bar struct {
	BarA int64
	BarB string
}

func TestComplexEnum(t *testing.T) {
	x := ComplexEnum{
		Enum: 0,
		Foo: Foo{
			FooA: 23,
			FooB: "baz",
		},
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(ComplexEnum)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)

	require.Equal(t, x, *y)
}

type S struct {
	S map[int]struct{}
}

func TestSet(t *testing.T) {
	x := S{
		S: map[int]struct{}{124: struct{}{}, 214: struct{}{}, 24: struct{}{}, 53: struct{}{}},
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(S)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

type Skipped struct {
	A int
	B int `borsh_skip:"true"`
	C int
}

func TestSkipped(t *testing.T) {
	x := Skipped{
		A: 32,
		B: 535,
		C: 123,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(Skipped)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)

	require.Equal(t, x.A, y.A)
	require.Equal(t, x.C, y.C)
	require.NotEqual(t, y.B, x.B, "didn't skip field B")
}

type E struct{}

func TestEmpty(t *testing.T) {
	x := E{}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)
	if len(data) != 0 {
		t.Error("not empty")
	}
	y := new(E)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)
	require.Equal(t, x, *y)
}

func testValue(t *testing.T, v interface{}) {
	data, err := MarshalBorsh(v)
	require.NoError(t, err)

	parsed := reflect.New(reflect.TypeOf(v))
	err = UnmarshalBorsh(parsed.Interface(), data)
	require.NoError(t, err)
	require.Equal(t, v, parsed.Elem().Interface())
}

func TestStrings(t *testing.T) {
	tests := []struct {
		in string
	}{
		{""},
		{"a"},
		{"hellow world"},
		{strings2.Repeat("x", 1024)},
		{strings2.Repeat("x", 4096)},
		{strings2.Repeat("x", 65535)},
		{strings2.Repeat("hello world!", 1000)},
		{"💩"},
	}

	for _, tt := range tests {
		testValue(t, tt.in)
	}
}

func makeInt32Slice(val int32, len int) []int32 {
	s := make([]int32, len)
	for i := 0; i < len; i++ {
		s[i] = val
	}
	return s
}

func TestSlices(t *testing.T) {
	tests := []struct {
		in []int32
	}{
		{nil}, // zero length slice
		{makeInt32Slice(1000000000, 1)},
		{makeInt32Slice(1000000001, 2)},
		{makeInt32Slice(1000000002, 3)},
		{makeInt32Slice(1000000003, 4)},
		{makeInt32Slice(1000000004, 8)},
		{makeInt32Slice(1000000005, 16)},
		{makeInt32Slice(1000000006, 32)},
		{makeInt32Slice(1000000007, 64)},
		{makeInt32Slice(1000000008, 65)},
	}

	for _, tt := range tests {
		testValue(t, tt.in)
	}
}

func TestUint128(t *testing.T) {
	tests := []struct {
		in Int128
	}{
		{func() Int128 {
			var v = Int128{
				Hi: math.MaxInt16,
				Lo: math.MaxInt16,
			}
			return v
		}()},
	}

	for _, tt := range tests {
		testValue(t, tt.in)
	}
}

type Myu8 uint8
type Myu16 uint16
type Myu32 uint32
type Myu64 uint64
type Myi8 int8
type Myi16 int16
type Myi32 int32
type Myi64 int64

type CustomType struct {
	U8  Myu8
	U16 Myu16
	U32 Myu32
	U64 Myu64
	I8  Myi8
	I16 Myi16
	I32 Myi32
	I64 Myi64
}

func TestCustomType(t *testing.T) {
	x := CustomType{
		U8:  1,
		U16: 2,
		U32: 3,
		U64: 4,
		I8:  5,
		I16: 6,
		I32: 7,
		I64: 8,
	}
	data, err := MarshalBorsh(x)
	require.NoError(t, err)

	y := new(CustomType)
	err = UnmarshalBorsh(y, data)
	require.NoError(t, err)

	require.Equal(t, x, *y)
}
