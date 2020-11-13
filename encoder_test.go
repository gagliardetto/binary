package bin

import (
	"bytes"
	"encoding/hex"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoder_int8(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	v := int8(-99)
	enc.writeByte(byte(v))
	enc.writeByte(byte(int8(100)))

	assert.Equal(t, []byte{
		0x9d, // -99
		0x64, // 100
	}, buf.Bytes())
}

func TestEncoder_int16(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeInt16(int16(-82))
	enc.writeInt16(int16(73))

	assert.Equal(t, []byte{
		0xae, 0xff, // -82
		0x49, 0x00, // 73
	}, buf.Bytes())
}

func TestEncoder_int32(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeInt32(int32(-276132392))
	enc.writeInt32(int32(237391))

	assert.Equal(t, []byte{
		0xd8, 0x8d, 0x8a, 0xef,
		0x4f, 0x9f, 0x3, 0x00,
	}, buf.Bytes())
}

func TestEncoder_int64(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeInt64(int64(-819823))
	enc.writeInt64(int64(72931))

	assert.Equal(t, []byte{
		0x91, 0x7d, 0xf3, 0xff, 0xff, 0xff, 0xff, 0xff, //-819823
		0xe3, 0x1c, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, //72931
	}, buf.Bytes())
}

func TestEncoder_uint8(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeByte(uint8(99))
	enc.writeByte(uint8(100))

	assert.Equal(t, []byte{
		0x63, // 99
		0x64, // 100
	}, buf.Bytes())
}

func TestEncoder_uint16(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeUint16(uint16(82))
	enc.writeUint16(uint16(73))

	assert.Equal(t, []byte{
		0x52, 0x00, // 82
		0x49, 0x00, // 73
	}, buf.Bytes())
}

func TestEncoder_uint32(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeUint32(uint32(276132392))
	enc.writeUint32(uint32(237391))

	assert.Equal(t, []byte{
		0x28, 0x72, 0x75, 0x10, // 276132392 as LE
		0x4f, 0x9f, 0x03, 0x00, // 237391 as LE
	}, buf.Bytes())
}

func TestEncoder_uint64(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeUint64(uint64(819823))
	enc.writeUint64(uint64(72931))

	assert.Equal(t, []byte{
		0x6f, 0x82, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, //819823
		0xe3, 0x1c, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, //72931
	}, buf.Bytes())
}

func TestEncoder_float32(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeFloat32(float32(1.32))
	enc.writeFloat32(float32(-3.21))

	assert.Equal(t, []byte{
		0xc3, 0xf5, 0xa8, 0x3f,
		0xa4, 0x70, 0x4d, 0xc0,
	}, buf.Bytes())
}

func TestEncoder_float64(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeFloat64(float64(-62.23))
	enc.writeFloat64(float64(23.239))
	enc.writeFloat64(float64(math.Inf(1)))
	enc.writeFloat64(float64(math.Inf(-1)))

	assert.Equal(t, []byte{
		0x3d, 0x0a, 0xd7, 0xa3, 0x70, 0x1d, 0x4f, 0xc0,
		0x77, 0xbe, 0x9f, 0x1a, 0x2f, 0x3d, 0x37, 0x40,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x7f,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xff,
	}, buf.Bytes())
}

func TestEncoder_string(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeString("123")
	enc.writeString("")
	enc.writeString("abc")

	assert.Equal(t, []byte{
		0x03, 0x31, 0x32, 0x33, // "123"
		0x00,                   // ""
		0x03, 0x61, 0x62, 0x63, // "abc
	}, buf.Bytes())
}

func TestEncoder_byte(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeByte(0)
	enc.writeByte(1)

	assert.Equal(t, []byte{
		0x00, 0x01,
	}, buf.Bytes())
}

func TestEncoder_bool(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeBool(true)
	enc.writeBool(false)

	assert.Equal(t, []byte{
		0x01, 0x00,
	}, buf.Bytes())
}

func TestEncoder_ByteArray(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.writeByteArray([]byte{1, 2, 3})
	enc.writeByteArray([]byte{4, 5, 6})

	assert.Equal(t, []byte{
		0x03, 0x01, 0x02, 0x03,
		0x03, 0x04, 0x05, 0x06,
	}, buf.Bytes())
}

func TestEncode_Array(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.Encode([3]byte{1, 2, 4})

	assert.Equal(t,
		[]byte{1, 2, 4},
		buf.Bytes(),
	)
}

func TestEncode_Array_Err(t *testing.T) {
	buf := new(bytes.Buffer)

	toEncode := [1]time.Duration{}

	enc := NewEncoder(buf)
	err := enc.Encode(toEncode)
	assert.EqualError(t, err, "encode: unsupported type time.Duration")
}

func TestEncoder_Slide_Err(t *testing.T) {

	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	err := enc.Encode([]time.Duration{time.Duration(0)})
	assert.EqualError(t, err, "encode: unsupported type time.Duration")

}

func Test_OptionalPointerToPrimitiveType(t *testing.T) {
	type test struct {
		ID *Uint64 `bin:"optional"`
	}

	out, err := MarshalBinary(test{ID: nil})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x0}, out)

	id := Uint64(0)
	out, err = MarshalBinary(test{ID: &id})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, out)

	id = Uint64(10)
	out, err = MarshalBinary(test{ID: &id})
	require.NoError(t, err)

	assert.Equal(t, []byte{0x1, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, out)
}

func TestEncoder_Int64(t *testing.T) {
	buf := new(bytes.Buffer)

	enc := NewEncoder(buf)
	enc.Encode(Int64(-819823))
	enc.Encode(Int64(72931))

	assert.Equal(t, []byte{
		0x91, 0x7d, 0xf3, 0xff, 0xff, 0xff, 0xff, 0xff, //-819823
		0xe3, 0x1c, 0x1, 0x00, 0x00, 0x00, 0x00, 0x00, //72931
	}, buf.Bytes())
}

func TestEncoder_BinaryStruct(t *testing.T) {
	s := &binaryTestStruct{
		F1:  "abc",
		F2:  -75,
		F3:  99,
		F4:  -231,
		F5:  999,
		F6:  -13231,
		F7:  99999,
		F8:  -23.13,
		F9:  3.92,
		F10: []string{"def", "789"},
		F11: [2]string{"foo", "bar"},
		F12: 0xff,
		F13: []byte{1, 2, 3, 4, 5},
		F14: true,
		F15: Int64(-23),
		F16: Uint64(23),
		F17: JSONFloat64(3.14),
		F18: Uint128{
			Lo: 10,
			Hi: 82,
		},
		F19: Int128{
			Lo: 7,
			Hi: 3,
		},
		F20: Float128{
			Lo: 10,
			Hi: 82,
		},
		F21: Varuint32(999),
		F22: Varint32(-999),
		F23: Bool(true),
		F24: HexBytes([]byte{1, 2, 3, 4, 5}),
	}
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	assert.NoError(t, enc.Encode(s))

	assert.Equal(t,
		"03616263b5ff630019ffffffe703000051ccffffffffffff9f860100000000003d0ab9c15c8fc2f5285c0f4002036465660337383903666f6f03626172ff05010203040501e9ffffffffffffff17000000000000001f85eb51b81e09400a000000000000005200000000000000070000000000000003000000000000000a000000000000005200000000000000e707cd0f01050102030405",
		hex.EncodeToString(buf.Bytes()),
	)
}

func TestEncoder_BinaryStruct_Err(t *testing.T) {
	s := binaryInvalidTestStruct{}

	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	err := enc.Encode(s)
	assert.EqualError(t, err, "encode: unsupported type time.Duration")
}
