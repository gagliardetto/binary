package bin

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"

	"go.uber.org/zap"
)

// UnmarshalerBinary is the interface implemented by types
// that can unmarshal an EOSIO binary description of themselves.
//
// **Warning** This is experimental, exposed only for internal usage for now.
type UnmarshalerBinary interface {
	UnmarshalBinary(decoder *Decoder) error
}

var TypeSize = struct {
	Bool int
	Byte int

	Int8  int
	Int16 int

	Uint8   int
	Uint16  int
	Uint32  int
	Uint64  int
	Uint128 int

	Float32 int
	Float64 int

	PublicKey int
	Signature int

	Tstamp         int
	BlockTimestamp int

	CurrencyName int
}{
	Byte: 1,
	Bool: 1,

	Int8:  1,
	Int16: 2,

	Uint8:   1,
	Uint16:  2,
	Uint32:  4,
	Uint64:  8,
	Uint128: 16,

	Float32: 4,
	Float64: 8,
}

// Decoder implements the EOS unpacking, similar to FC_BUFFER
type Decoder struct {
	data             []byte
	pos              int
	decodeP2PMessage bool
	decodeActions    bool
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data:             data,
		decodeP2PMessage: true,
		decodeActions:    true,
	}
}

func (d *Decoder) Decode(v interface{}) (err error) {
	return d.DecodeWithOption(v, nil)
}

func (d *Decoder) DecodeWithOption(v interface{}, option *Option) (err error) {
	if option == nil {
		option = &Option{}
	}

	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.CanAddr() {
		return fmt.Errorf("can only decode to pointer type, got %T", v)
	}
	t := rv.Type()

	if traceEnabled {
		zlog.Debug("decode type",
			typeField("type", v),
			zap.Reflect("options", option),
		)
	}

	if option.isOptional() {
		isPresent, e := d.ReadByte()
		if e != nil {
			err = fmt.Errorf("decode: %t isPresent, %s", v, e)
			return
		}

		if isPresent == 0 {
			if traceEnabled {
				zlog.Debug("skipping optional", typeField("type", v))
			}

			rv.Set(reflect.Zero(t))
			return
		}
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		newRV := reflect.New(t)
		rv.Set(newRV)

		// At this point, `newRV` is a pointer to our target type, we need to check here because
		// after that, when `reflect.Indirect` is used, we get a `**<Type>` instead of a `*<Type>`
		// which breaks the interface checking.
		//
		// Ultimately, I think this could should be re-written, I don't think the `**<Type>` is necessary.
		if u, ok := newRV.Interface().(UnmarshalerBinary); ok {
			if traceEnabled {
				zlog.Debug("using UnmarshalBinary method to decode type", typeField("type", v))
			}
			return u.UnmarshalBinary(d)
		}

		rv = reflect.Indirect(newRV)
	} else {
		// We check if `v` directly is `UnmarshalerBinary` this is to overcome our bad code that
		// has problem dealing with non-pointer type, which should still be possible here, by allocating
		// the empty value for it can then unmarshalling using the address of it. See comment above about
		// `newRV` being turned into `**<Type>`.
		//
		// We should re-code all the logic to determine the type and indirection using Golang `json` package
		// logic. See here: https://github.com/golang/go/blob/54697702e435bddb69c0b76b25b3209c78d2120a/src/encoding/json/decode.go#L439
		if u, ok := v.(UnmarshalerBinary); ok {
			if traceEnabled {
				zlog.Debug("using UnmarshalBinary method to decode type", typeField("type", v))
			}
			return u.UnmarshalBinary(d)
		}
	}

	switch v.(type) {
	case *string:
		s, e := d.ReadString()
		if e != nil {
			err = e
			return
		}
		rv.SetString(s)
		return
	case *byte:
		var n byte
		n, err = d.ReadByte()
		rv.SetUint(uint64(n))
		return
	case *int8:
		var n int8
		n, err = d.ReadInt8()
		rv.SetInt(int64(n))
		return
	case *int16:
		var n int16
		n, err = d.ReadInt16()
		rv.SetInt(int64(n))
		return
	case *int32:
		var n int32
		n, err = d.ReadInt32()
		rv.SetInt(int64(n))
		return
	case *int64:
		var n int64
		n, err = d.ReadInt64()
		rv.SetInt(int64(n))
		return
	// This is so hackish, doing it right now, but the decoder needs to handle those
	// case (a struct field that is itself a pointer) by itself.
	case **Uint64:
		var n uint64
		n, err = d.ReadUint64()
		if err == nil {
			rv.Set(reflect.ValueOf((Uint64)(n)))
		}
		return
	case *uint16:
		var n uint16
		n, err = d.ReadUint16()
		rv.SetUint(uint64(n))
		return
	case *uint32:
		var n uint32
		n, err = d.ReadUint32()
		rv.SetUint(uint64(n))
		return
	case *uint64:
		var n uint64
		n, err = d.ReadUint64()
		rv.SetUint(n)
		return
	case *Varint16:
		var r int16
		r, err = d.ReadVarint16()
		rv.SetInt(int64(r))
		return
	case *Varint32:
		var r int64
		r, err = d.ReadVarint64()
		rv.SetInt(r)
		return
	case *Varuint16:
		var r uint16
		r, err = d.ReadUvarint16()
		rv.SetUint(uint64(r))
		return
	case *float32:
		var n float32
		n, err = d.ReadFloat32()
		rv.SetFloat(float64(n))
		return
	case *float64:
		var n float64
		n, err = d.ReadFloat64()
		rv.SetFloat(n)
		return
	case *bool:
		var r bool
		r, err = d.ReadBool()
		rv.SetBool(r)
		return
	case *[]byte:
		var data []byte
		data, err = d.ReadByteArray()
		rv.SetBytes(data)
		return
	}

	switch t.Kind() {
	case reflect.Array:
		if traceEnabled {
			zlog.Debug("reading array")
		}
		len := t.Len()
		for i := 0; i < len; i++ {
			if err = d.DecodeWithOption(rv.Index(i).Addr().Interface(), nil); err != nil {
				return
			}
		}
		return

	case reflect.Slice:
		var l int
		if option.hasSizeOfSlice() {
			l = option.getSizeOfSlice()
		} else {
			length, err := d.ReadUvarint64()
			if err != nil {
				return err
			}
			l = int(length)
		}
		if traceEnabled {
			zlog.Debug("reading slice", zap.Int("len", l), typeField("type", v))
		}
		rv.Set(reflect.MakeSlice(t, int(l), int(l)))
		for i := 0; i < int(l); i++ {
			if err = d.DecodeWithOption(rv.Index(i).Addr().Interface(), nil); err != nil {
				return
			}
		}

	case reflect.Struct:

		err = d.decodeStruct(v, t, rv)
		if err != nil {
			return
		}

	default:
		return errors.New("decode: unsupported type " + t.String())
	}

	return
}

// rv is the instance of the structure
// t is the type of the structure
func (d *Decoder) decodeStruct(v interface{}, t reflect.Type, rv reflect.Value) (err error) {
	l := rv.NumField()

	sizeOfMap := map[string]int{}
	seenBinaryExtensionField := false
	for i := 0; i < l; i++ {
		structField := t.Field(i)

		fieldTag := parseFieldTag(structField.Tag)
		if fieldTag.Skip {
			continue
		}

		if !fieldTag.BinaryExtension && seenBinaryExtensionField {
			panic(fmt.Sprintf("the `bin:\"binary_extension\"` tags must be packed together at the end of struct fields, problematic field %q", structField.Name))
		}

		if fieldTag.BinaryExtension {
			seenBinaryExtensionField = true
			// FIXME: This works only if what is in `d.data` is the actual full data buffer that
			//        needs to be decoded. If there is for example two structs in the buffer, this
			//        will not work as we would continue into the next struct.
			//
			//        But at the same time, does it make sense otherwise? What would be the inference
			//        rule in the case of extra bytes available? Continue decoding and revert if it's
			//        not working? But how to detect valid errors?
			if len(d.data[d.pos:]) <= 0 {
				continue
			}
		}

		if v := rv.Field(i); v.CanSet() && structField.Name != "_" {
			option := &Option{}

			if s, ok := sizeOfMap[structField.Name]; ok {
				option.setSizeOfSlice(s)
			}

			// v is Value of given field for said struct
			if fieldTag.Optional {
				option.OptionalField = true
			}

			// creates a pointer to the value.....
			value := v.Addr().Interface()

			if traceEnabled {
				zlog.Debug("struct field",
					typeField(structField.Name, value),
					zap.Reflect("field_tags", fieldTag),
				)
			}

			if err = d.DecodeWithOption(value, option); err != nil {
				return
			}

			if fieldTag.Sizeof != "" {
				size := sizeof(structField.Type, v)
				if traceEnabled {
					zlog.Debug("setting size of field",
						zap.String("field_name", fieldTag.Sizeof),
						zap.Int("size", size),
					)
				}
				sizeOfMap[fieldTag.Sizeof] = size
			}
		}
	}
	return
}

func sizeof(t reflect.Type, v reflect.Value) int {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := int(v.Uint())
		// all the builtin array length types are native int
		// so this guards against weird truncation
		if n < 0 {
			return 0
		}
		return n
	default:
		//name := v.Type().FieldByIndex(index).Name
		//panic(fmt.Sprintf("sizeof field %T.%s not an integer type", val.Interface(), name))
		panic(fmt.Sprintf("sizeof field "))
	}
}

var ErrVarIntBufferSize = errors.New("varint: invalid buffer size")

func (d *Decoder) ReadUvarint64() (uint64, error) {
	l, read := binary.Uvarint(d.data[d.pos:])
	if read <= 0 {
		return l, ErrVarIntBufferSize
	}
	if traceEnabled {
		zlog.Debug("read uvarint64", zap.Uint64("val", l))
	}
	d.pos += read
	return l, nil
}

func (d *Decoder) ReadVarint64() (out int64, err error) {
	l, read := binary.Varint(d.data[d.pos:])
	if read <= 0 {
		return l, ErrVarIntBufferSize
	}
	if traceEnabled {
		zlog.Debug("read varint", zap.Int64("val", l))
	}
	d.pos += read
	return l, nil
}

func (d *Decoder) ReadVarint32() (out int32, err error) {
	n, err := d.ReadVarint64()
	if err != nil {
		return out, err
	}
	out = int32(n)
	if traceEnabled {
		zlog.Debug("read varint32", zap.Int32("val", out))
	}
	return
}

func (d *Decoder) ReadUvarint32() (out uint32, err error) {

	n, err := d.ReadUvarint64()
	if err != nil {
		return out, err
	}
	out = uint32(n)
	if traceEnabled {
		zlog.Debug("read uvarint32", zap.Uint32("val", out))
	}
	return
}
func (d *Decoder) ReadVarint16() (out int16, err error) {
	n, err := d.ReadVarint64()
	if err != nil {
		return out, err
	}
	out = int16(n)
	if traceEnabled {
		zlog.Debug("read varint16", zap.Int16("val", out))
	}
	return
}

func (d *Decoder) ReadUvarint16() (out uint16, err error) {

	n, err := d.ReadUvarint64()
	if err != nil {
		return out, err
	}
	out = uint16(n)
	if traceEnabled {
		zlog.Debug("read uvarint16", zap.Uint16("val", out))
	}
	return
}

func (d *Decoder) ReadByteArray() (out []byte, err error) {

	l, err := d.ReadUvarint64()
	if err != nil {
		return nil, err
	}

	if len(d.data) < d.pos+int(l) {
		return nil, fmt.Errorf("byte array: varlen=%d, missing %d bytes", l, d.pos+int(l)-len(d.data))
	}

	out = d.data[d.pos : d.pos+int(l)]
	d.pos += int(l)
	if traceEnabled {
		zlog.Debug("read byte array", zap.Stringer("hex", HexBytes(out)))
	}
	return
}

func (d *Decoder) ReadByte() (out byte, err error) {
	if d.remaining() < TypeSize.Byte {
		err = fmt.Errorf("required [1] byte, remaining [%d]", d.remaining())
		return
	}

	out = d.data[d.pos]
	d.pos++
	if traceEnabled {
		zlog.Debug("read byte", zap.Uint8("byte", out), zap.String("hex", hex.EncodeToString([]byte{out})))
	}
	return
}

func (d *Decoder) ReadBool() (out bool, err error) {
	if d.remaining() < TypeSize.Bool {
		err = fmt.Errorf("bool required [%d] byte, remaining [%d]", TypeSize.Bool, d.remaining())
		return
	}

	b, err := d.ReadByte()

	if err != nil {
		err = fmt.Errorf("readBool, %s", err)
	}
	out = b != 0
	if traceEnabled {
		zlog.Debug("read bool", zap.Bool("val", out))
	}
	return

}

func (d *Decoder) ReadUint8() (out uint8, err error) {
	out, err = d.ReadByte()
	return
}

func (d *Decoder) ReadInt8() (out int8, err error) {
	b, err := d.ReadByte()
	out = int8(b)
	if traceEnabled {
		zlog.Debug("read int8", zap.Int8("val", out))
	}
	return
}

func (d *Decoder) ReadUint16() (out uint16, err error) {
	if d.remaining() < TypeSize.Uint16 {
		err = fmt.Errorf("uint16 required [%d] bytes, remaining [%d]", TypeSize.Uint16, d.remaining())
		return
	}

	out = binary.LittleEndian.Uint16(d.data[d.pos:])
	d.pos += TypeSize.Uint16
	if traceEnabled {
		zlog.Debug("read uint16", zap.Uint16("val", out))
	}
	return
}

func (d *Decoder) ReadInt16() (out int16, err error) {
	n, err := d.ReadUint16()
	out = int16(n)
	if traceEnabled {
		zlog.Debug("read int16", zap.Int16("val", out))
	}
	return
}

func (d *Decoder) ReadInt64() (out int64, err error) {
	n, err := d.ReadUint64()
	out = int64(n)
	if traceEnabled {
		zlog.Debug("read int64", zap.Int64("val", out))
	}
	return
}

func (d *Decoder) ReadUint32() (out uint32, err error) {
	if d.remaining() < TypeSize.Uint32 {
		err = fmt.Errorf("uint32 required [%d] bytes, remaining [%d]", TypeSize.Uint32, d.remaining())
		return
	}

	out = binary.LittleEndian.Uint32(d.data[d.pos:])
	d.pos += TypeSize.Uint32
	if traceEnabled {
		zlog.Debug("read uint32", zap.Uint32("val", out))
	}
	return
}

func (d *Decoder) ReadInt32() (out int32, err error) {
	n, err := d.ReadUint32()
	out = int32(n)
	if traceEnabled {
		zlog.Debug("read int32", zap.Int32("val", out))
	}
	return
}

func (d *Decoder) ReadUint64() (out uint64, err error) {
	if d.remaining() < TypeSize.Uint64 {
		err = fmt.Errorf("uint64 required [%d] bytes, remaining [%d]", TypeSize.Uint64, d.remaining())
		return
	}

	data := d.data[d.pos : d.pos+TypeSize.Uint64]
	out = binary.LittleEndian.Uint64(data)
	d.pos += TypeSize.Uint64
	if traceEnabled {
		zlog.Debug("read uint64", zap.Uint64("val", out), zap.Stringer("hex", HexBytes(data)))
	}
	return
}

func (d *Decoder) ReadInt128() (out Int128, err error) {
	v, err := d.ReadUint128("int128")
	if err != nil {
		return
	}

	return Int128(v), nil
}

func (d *Decoder) ReadUint128(typeName string) (out Uint128, err error) {
	if d.remaining() < TypeSize.Uint128 {
		err = fmt.Errorf("%s required [%d] bytes, remaining [%d]", typeName, TypeSize.Uint128, d.remaining())
		return
	}

	data := d.data[d.pos : d.pos+TypeSize.Uint128]
	out.Lo = binary.LittleEndian.Uint64(data)
	out.Hi = binary.LittleEndian.Uint64(data[8:])

	d.pos += TypeSize.Uint128
	if traceEnabled {
		zlog.Debug("read uint128", zap.Stringer("hex", out), zap.Uint64("hi", out.Hi), zap.Uint64("lo", out.Lo))
	}
	return
}

func (d *Decoder) ReadFloat32() (out float32, err error) {
	if d.remaining() < TypeSize.Float32 {
		err = fmt.Errorf("float32 required [%d] bytes, remaining [%d]", TypeSize.Float32, d.remaining())
		return
	}

	n := binary.LittleEndian.Uint32(d.data[d.pos:])
	out = math.Float32frombits(n)
	d.pos += TypeSize.Float32
	if traceEnabled {
		zlog.Debug("read float32", zap.Float32("val", out))
	}
	return
}

func (d *Decoder) ReadFloat64() (out float64, err error) {
	if d.remaining() < TypeSize.Float64 {
		err = fmt.Errorf("float64 required [%d] bytes, remaining [%d]", TypeSize.Float64, d.remaining())
		return
	}

	n := binary.LittleEndian.Uint64(d.data[d.pos:])
	out = math.Float64frombits(n)
	d.pos += TypeSize.Float64
	if traceEnabled {
		zlog.Debug("read Float64", zap.Float64("val", float64(out)))
	}
	return
}

func (d *Decoder) ReadFloat128() (out Float128, err error) {
	value, err := d.ReadUint128("float128")
	if err != nil {
		return out, fmt.Errorf("float128: %s", err)
	}

	return Float128(value), nil
}

func (d *Decoder) SafeReadUTF8String() (out string, err error) {
	data, err := d.ReadByteArray()
	out = strings.Map(fixUtf, string(data))
	if traceEnabled {
		zlog.Debug("read safe UTF8 string", zap.String("val", out))
	}
	return
}

func (d *Decoder) ReadString() (out string, err error) {
	data, err := d.ReadByteArray()
	out = string(data)
	if traceEnabled {
		zlog.Debug("read string", zap.String("val", out))
	}
	return
}

func (d *Decoder) remaining() int {
	return len(d.data) - d.pos
}

func (d *Decoder) hasRemaining() bool {
	return d.remaining() > 0
}

//func UnmarshalBinaryReader(reader io.Reader, v interface{}) (err error) {
//	data, err := ioutil.ReadAll(reader)
//	if err != nil {
//		return
//	}
//	return UnmarshalBinary(data, v)
//}
//
//func UnmarshalBinary(data []byte, v interface{}) (err error) {
//	decoder := NewDecoder(data)
//	return decoder.Decode(v)
//}

func fixUtf(r rune) rune {
	if r == utf8.RuneError {
		return '�'
	}
	return r
}
