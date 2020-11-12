package bin

import (
	"encoding/binary"
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

func (d *Decoder) DecodeP2PMessage(decode bool) {
	d.decodeP2PMessage = decode
}

func (d *Decoder) DecodeActions(decode bool) {
	d.decodeActions = decode
}

type DecodeOption = interface{}

type optionalFieldType bool

const OptionalField optionalFieldType = true

func (d *Decoder) Decode(v interface{}, options ...DecodeOption) (err error) {
	optionalField := false
	for _, option := range options {
		if _, isOptionalField := option.(optionalFieldType); isOptionalField {
			optionalField = true
		}
	}

	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.CanAddr() {
		return fmt.Errorf("can only decode to pointer type, got %T", v)
	}
	t := rv.Type()

	if loggingEnabled {
		decoderLog.Debug("decode type", typeField("type", v), zap.Bool("optional", optionalField))
	}

	if optionalField {
		isPresent, e := d.ReadByte()
		if e != nil {
			err = fmt.Errorf("decode: %t isPresent, %s", v, e)
			return
		}

		if isPresent == 0 {
			if loggingEnabled {
				decoderLog.Debug("skipping optional", typeField("type", v))
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
			if loggingEnabled {
				decoderLog.Debug("using UnmarshalBinary method to decode type", typeField("type", v))
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
			if loggingEnabled {
				decoderLog.Debug("using UnmarshalBinary method to decode type", typeField("type", v))
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
	case *Int64:
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
	case *Uint64:
		var n uint64
		n, err = d.ReadUint64()
		rv.SetUint(uint64(n))
		return
	case *JSONFloat64:
		var n float64
		n, err = d.ReadFloat64()
		rv.SetFloat(n)
		return
	case *Uint128:
		var n Uint128
		n, err = d.ReadUint128("uint128")
		rv.Set(reflect.ValueOf(n))
		return
	case *Int128:
		var n Uint128
		n, err = d.ReadUint128("int128")
		rv.Set(reflect.ValueOf(Int128(n)))
		return
	case *Float128:
		var n Uint128
		n, err = d.ReadUint128("float128")
		rv.Set(reflect.ValueOf(Float128(n)))
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
	case *Varuint32:
		var r uint64
		r, err = d.ReadUvarint64()
		rv.SetUint(r)
		return
	case *bool:
		var r bool
		r, err = d.ReadBool()
		rv.SetBool(r)
		return
	case *Bool:
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
		if loggingEnabled {
			decoderLog.Debug("reading array")
		}
		len := t.Len()
		for i := 0; i < int(len); i++ {
			if err = d.Decode(rv.Index(i).Addr().Interface()); err != nil {
				return
			}
		}
		return

	case reflect.Slice:
		var l uint64
		if l, err = d.ReadUvarint64(); err != nil {
			return
		}
		if loggingEnabled {
			decoderLog.Debug("reading slice", zap.Uint64("len", l), typeField("type", v))
		}
		rv.Set(reflect.MakeSlice(t, int(l), int(l)))
		for i := 0; i < int(l); i++ {
			if err = d.Decode(rv.Index(i).Addr().Interface()); err != nil {
				return
			}
		}

	case reflect.Struct:

		err = d.decodeStruct(v, t, rv)
		if err != nil {
			return
		}

	default:
		return errors.New("decode, unsupported type " + t.String())
	}

	return
}

func (d *Decoder) decodeStruct(v interface{}, t reflect.Type, rv reflect.Value) (err error) {
	l := rv.NumField()

	seenBinaryExtensionField := false
	for i := 0; i < l; i++ {
		structField := t.Field(i)
		tag := structField.Tag.Get("eos")
		if tag == "-" {
			continue
		}

		if tag != "binary_extension" && seenBinaryExtensionField {
			panic(fmt.Sprintf("the `eos: \"binary_extension\"` tags must be packed together at the end of struct fields, problematic field %s", structField.Name))
		}

		if tag == "binary_extension" {
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
			var options []DecodeOption
			if tag == "optional" {
				options = append(options, OptionalField)
			}

			value := v.Addr().Interface()

			if loggingEnabled {
				decoderLog.Debug("struct field", typeField(structField.Name, value), zap.String("tag", tag))
			}

			if err = d.Decode(value, options...); err != nil {
				return
			}
		}
	}
	return
}

var ErrVarIntBufferSize = errors.New("varint: invalid buffer size")

func (d *Decoder) ReadUvarint64() (uint64, error) {
	l, read := binary.Uvarint(d.data[d.pos:])
	if read <= 0 {
		return l, ErrVarIntBufferSize
	}
	if loggingEnabled {
		decoderLog.Debug("read uvarint64", zap.Uint64("val", l))
	}
	d.pos += read
	return l, nil
}

func (d *Decoder) ReadVarint64() (out int64, err error) {
	l, read := binary.Varint(d.data[d.pos:])
	if read <= 0 {
		return l, ErrVarIntBufferSize
	}
	if loggingEnabled {
		decoderLog.Debug("read varint", zap.Int64("val", l))
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
	if loggingEnabled {
		decoderLog.Debug("read varint32", zap.Int32("val", out))
	}
	return
}

func (d *Decoder) ReadUvarint32() (out uint32, err error) {

	n, err := d.ReadUvarint64()
	if err != nil {
		return out, err
	}
	out = uint32(n)
	if loggingEnabled {
		decoderLog.Debug("read uvarint32", zap.Uint32("val", out))
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
	if loggingEnabled {
		decoderLog.Debug("read byte array", zap.Stringer("hex", HexBytes(out)))
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
	if loggingEnabled {
		decoderLog.Debug("read byte", zap.Uint8("byte", out))
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
	if loggingEnabled {
		decoderLog.Debug("read bool", zap.Bool("val", out))
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
	if loggingEnabled {
		decoderLog.Debug("read int8", zap.Int8("val", out))
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
	if loggingEnabled {
		decoderLog.Debug("read uint16", zap.Uint16("val", out))
	}
	return
}

func (d *Decoder) ReadInt16() (out int16, err error) {
	n, err := d.ReadUint16()
	out = int16(n)
	if loggingEnabled {
		decoderLog.Debug("read int16", zap.Int16("val", out))
	}
	return
}

func (d *Decoder) ReadInt64() (out int64, err error) {
	n, err := d.ReadUint64()
	out = int64(n)
	if loggingEnabled {
		decoderLog.Debug("read int64", zap.Int64("val", out))
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
	if loggingEnabled {
		decoderLog.Debug("read uint32", zap.Uint32("val", out))
	}
	return
}

func (d *Decoder) ReadInt32() (out int32, err error) {
	n, err := d.ReadUint32()
	out = int32(n)
	if loggingEnabled {
		decoderLog.Debug("read int32", zap.Int32("val", out))
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
	if loggingEnabled {
		decoderLog.Debug("read uint64", zap.Uint64("val", out), zap.Stringer("hex", HexBytes(data)))
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
	if loggingEnabled {
		decoderLog.Debug("read uint128", zap.Stringer("hex", out), zap.Uint64("hi", out.Hi), zap.Uint64("lo", out.Lo))
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
	if loggingEnabled {
		decoderLog.Debug("read float32", zap.Float32("val", out))
	}
	return
}

func (d *Decoder) ReadNodeosFloat32() (out float32, err error) {
	if d.remaining() < TypeSize.Float32 {
		err = fmt.Errorf("float32 required [%d] bytes, remaining [%d]", TypeSize.Float32, d.remaining())
		return
	}

	n := binary.LittleEndian.Uint32(d.data[d.pos:])
	out = math.Float32frombits(n)
	d.pos += TypeSize.Float32
	if loggingEnabled {
		decoderLog.Debug("read float32", zap.Float32("val", out))
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
	if loggingEnabled {
		decoderLog.Debug("read Float64", zap.Float64("val", float64(out)))
	}
	return
}

func (d *Decoder) SafeReadUTF8String() (out string, err error) {
	data, err := d.ReadByteArray()
	out = strings.Map(fixUtf, string(data))
	if loggingEnabled {
		decoderLog.Debug("read safe UTF8 string", zap.String("val", out))
	}
	return
}

func (d *Decoder) ReadString() (out string, err error) {
	data, err := d.ReadByteArray()
	out = string(data)
	if loggingEnabled {
		decoderLog.Debug("read string", zap.String("val", out))
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
