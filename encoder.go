package bin

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
	"strings"

	"go.uber.org/zap"
)

type Encoder struct {
	output io.Writer
	Order  binary.ByteOrder
	count  int
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		output: w,
		Order:  binary.LittleEndian,
		count:  0,
	}
}

func (e *Encoder) Encode(v interface{}) (err error) {
	return e.EncodeWithOption(v, nil)
}
func (e *Encoder) EncodeWithOption(v interface{}, option *Option) (err error) {
	if option == nil {
		option = &Option{}
	}

	switch cv := v.(type) {
	case MarshalerBinary:
		return cv.MarshalBinary(e)
	case BaseVariant:

		err = e.WriteUVarInt(int(cv.TypeID))
		if err != nil {
			return
		}
		return e.Encode(cv.Impl)
	case SafeString:
		return e.WriteString(string(cv))
	case string:
		return e.WriteString(cv)
	case byte:
		return e.WriteByte(cv)
	case int8:
		return e.WriteByte(byte(cv))
	case int16:
		return e.WriteInt16(cv)
	case uint16:
		return e.WriteUint16(cv)
	case int32:
		return e.WriteInt32(cv)
	case uint32:
		return e.WriteUint32(cv)
	case uint64:
		return e.WriteUint64(cv)
	case int64:
		return e.WriteInt64(cv)
	case float32:
		return e.WriteFloat32(cv)
	case float64:
		return e.WriteFloat64(cv)
	case bool:
		return e.WriteBool(cv)
	case []byte:
		// if the byte array has another filed
		// that specifies it size we should not prefix
		// the data with the length of the byte array
		writeLength := !option.hasSizeOfSlice()
		return e.WriteByteArray(cv, writeLength)
	case nil:
	default:

		rv := reflect.Indirect(reflect.ValueOf(v))
		t := rv.Type()

		switch t.Kind() {

		case reflect.Array:
			l := t.Len()

			if traceEnabled {
				defer func(prev *zap.Logger) { zlog = prev }(zlog)
				zlog = zlog.Named("array")
				zlog.Debug("encode: array", zap.Int("length", l), typeField("type", v))
			}

			for i := 0; i < l; i++ {
				if err = e.Encode(rv.Index(i).Interface()); err != nil {
					return
				}
			}
		case reflect.Slice:
			var l int
			if option.hasSizeOfSlice() {
				l = option.getSizeOfSlice()
				if traceEnabled {
					zlog.Debug("slice with size of", zap.Int("size_of", l))
				}
			} else {
				l = rv.Len()
				if err = e.WriteUVarInt(l); err != nil {
					return
				}
			}

			if traceEnabled {
				defer func(prev *zap.Logger) { zlog = prev }(zlog)
				zlog = zlog.Named("slice")
				zlog.Debug("encode: slice", zap.Int("length", l), typeField("type", v))
			}

			for i := 0; i < l; i++ {
				if err = e.Encode(rv.Index(i).Interface()); err != nil {
					return
				}
			}
		case reflect.Struct:
			l := rv.NumField()

			if traceEnabled {
				zlog.Debug("encode: struct", zap.Int("fields", l), typeField("type", v))
				defer func(prev *zap.Logger) { zlog = prev }(zlog)
				zlog = zlog.Named("struct")
			}

			sizeOfMap := map[string]int{}

			for i := 0; i < l; i++ {
				field := t.Field(i)

				tag := field.Tag.Get("bin")
				if tag == "-" {
					continue
				}
				lookup := "sizeof=" + field.Name
				if s, ok := sizeOfMap[lookup]; ok {
					if traceEnabled {
						zlog.Debug("setting sizeof option", zap.String("of", lookup), zap.Int("size", s))
					}
					option.setSizeOfSlice(s)
				}

				if traceEnabled {
					zlog.Debug("field", zap.String("field", field.Name))
				}

				if v := rv.Field(i); t.Field(i).Name != "_" {
					if strings.HasPrefix(tag, "sizeof=") {
						if traceEnabled {
							zlog.Debug("found sizeof", zap.String("size_of", tag))
						}
						sizeOfMap[tag] = sizeof(field.Type, v)
					}
					if v.CanInterface() {
						isPresent := true
						if tag == "optional" {
							isPresent = !v.IsZero()
							e.WriteBool(isPresent)
						}

						if isPresent {
							if err = e.EncodeWithOption(v.Interface(), option); err != nil {
								return
							}
						}
					}
				}
			}

		case reflect.Map:
			keyCount := len(rv.MapKeys())

			if traceEnabled {
				zlog.Debug("encode: map", zap.Int("key_count", keyCount), typeField("key_type", t.Key()), typeField("value_type", rv.Elem()))
				defer func(prev *zap.Logger) { zlog = prev }(zlog)
				zlog = zlog.Named("struct")
			}

			if err = e.WriteUVarInt(keyCount); err != nil {
				return
			}

			for _, mapKey := range rv.MapKeys() {
				if err = e.Encode(mapKey.Interface()); err != nil {
					return
				}

				if err = e.Encode(rv.MapIndex(mapKey).Interface()); err != nil {
					return
				}
			}

		default:
			return errors.New("encode: unsupported type " + t.String())
		}
	}

	return
}

func (e *Encoder) toWriter(bytes []byte) (err error) {
	e.count += len(bytes)

	if traceEnabled {
		zlog.Debug("    appending", zap.Stringer("hex", HexBytes(bytes)), zap.Int("pos", e.count))
	}

	_, err = e.output.Write(bytes)
	return
}

func (e *Encoder) WriteByteArray(b []byte, writeLength bool) error {
	if traceEnabled {
		zlog.Debug("write byte array", zap.Int("len", len(b)))
	}
	if writeLength {
		if err := e.WriteUVarInt(len(b)); err != nil {
			return err
		}
	}
	return e.toWriter(b)
}

func (e *Encoder) WriteUVarInt(v int) (err error) {
	if traceEnabled {
		zlog.Debug("write uvarint", zap.Int("val", v))
	}

	buf := make([]byte, 8)
	l := binary.PutUvarint(buf, uint64(v))
	return e.toWriter(buf[:l])
}

func (e *Encoder) WriteVarInt(v int) (err error) {
	if traceEnabled {
		zlog.Debug("write varint", zap.Int("val", v))
	}

	buf := make([]byte, 8)
	l := binary.PutVarint(buf, int64(v))
	return e.toWriter(buf[:l])
}

func (e *Encoder) WriteByte(b byte) (err error) {
	if traceEnabled {
		zlog.Debug("write byte", zap.Uint8("val", b))
	}
	return e.toWriter([]byte{b})
}

func (e *Encoder) WriteBool(b bool) (err error) {
	if traceEnabled {
		zlog.Debug("write bool", zap.Bool("val", b))
	}
	var out byte
	if b {
		out = 1
	}
	return e.WriteByte(out)
}

func (e *Encoder) WriteUint8(i uint8) (err error) {
	return e.WriteByte(i)
}

func (e *Encoder) WriteUint16(i uint16) (err error) {
	if traceEnabled {
		zlog.Debug("write uint16", zap.Uint16("val", i))
	}
	buf := make([]byte, TypeSize.Uint16)
	binary.LittleEndian.PutUint16(buf, i)
	return e.toWriter(buf)
}

func (e *Encoder) WriteInt16(i int16) (err error) {
	if traceEnabled {
		zlog.Debug("write int16", zap.Int16("val", i))
	}
	return e.WriteUint16(uint16(i))
}

func (e *Encoder) WriteInt32(i int32) (err error) {
	if traceEnabled {
		zlog.Debug("write int32", zap.Int32("val", i))
	}
	return e.WriteUint32(uint32(i))
}

//WriteUint32
func (e *Encoder) WriteUint32(i uint32) (err error) {
	if traceEnabled {
		zlog.Debug("write uint32", zap.Uint32("val", i))
	}
	buf := make([]byte, TypeSize.Uint32)
	binary.LittleEndian.PutUint32(buf, i)
	return e.toWriter(buf)
}

func (e *Encoder) WriteInt64(i int64) (err error) {
	if traceEnabled {
		zlog.Debug("write int64", zap.Int64("val", i))
	}
	return e.WriteUint64(uint64(i))
}

func (e *Encoder) WriteUint64(i uint64) (err error) {
	if traceEnabled {
		zlog.Debug("write uint64", zap.Uint64("val", i))
	}
	buf := make([]byte, TypeSize.Uint64)
	binary.LittleEndian.PutUint64(buf, i)
	return e.toWriter(buf)
}

func (e *Encoder) WriteUint128(i Uint128) (err error) {
	if traceEnabled {
		zlog.Debug("write uint128", zap.Stringer("hex", i), zap.Uint64("lo", i.Lo), zap.Uint64("hi", i.Hi))
	}
	buf := make([]byte, TypeSize.Uint128)
	binary.LittleEndian.PutUint64(buf, i.Lo)
	binary.LittleEndian.PutUint64(buf[TypeSize.Uint64:], i.Hi)
	return e.toWriter(buf)
}

func (e *Encoder) WriteInt128(i Int128) (err error) {
	if traceEnabled {
		zlog.Debug("write int128", zap.Stringer("hex", i), zap.Uint64("lo", i.Lo), zap.Uint64("hi", i.Hi))
	}
	buf := make([]byte, TypeSize.Uint128)
	binary.LittleEndian.PutUint64(buf, i.Lo)
	binary.LittleEndian.PutUint64(buf[TypeSize.Uint64:], i.Hi)
	return e.toWriter(buf)
}

func (e *Encoder) WriteFloat32(f float32) (err error) {
	if traceEnabled {
		zlog.Debug("write float32", zap.Float32("val", f))
	}
	i := math.Float32bits(f)
	buf := make([]byte, TypeSize.Uint32)
	binary.LittleEndian.PutUint32(buf, i)

	return e.toWriter(buf)
}
func (e *Encoder) WriteFloat64(f float64) (err error) {
	if traceEnabled {
		zlog.Debug("write float64", zap.Float64("val", f))
	}
	i := math.Float64bits(f)
	buf := make([]byte, TypeSize.Uint64)
	binary.LittleEndian.PutUint64(buf, i)

	return e.toWriter(buf)
}

func (e *Encoder) WriteString(s string) (err error) {
	if traceEnabled {
		zlog.Debug("write string", zap.String("val", s))
	}
	return e.WriteByteArray([]byte(s), true)
}
