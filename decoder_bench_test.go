package bin

import (
	"reflect"
	"testing"
)

func Benchmark_uintSlices_Decode(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		// length:
		uint32ToBytes(uint32(l), LE),
		// data:
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := make([]uint32, 0)
		decoder := NewBorshDecoder(buf)
		err := decoder.Decode(&got)
		if err != nil {
			b.Error(err)
		}
	}
}

func newUint64SliceEncoded(l int) []byte {
	buf := make([]byte, 0)
	for i := 0; i < l; i++ {
		buf = append(buf, uint64ToBytes(uint64(i), LE)...)
	}
	return buf
}

func Benchmark_uintSlices_append(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder := NewBorshDecoder(buf)

		var got []uint64
		rv := reflect.ValueOf(&got).Elem()
		k := rv.Type().Elem().Kind()

		err := reflect_readArrayOfUint_(decoder, len(buf)/8, k, rv, LE)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}

func Benchmark_uintSlices_preallocate(b *testing.B) {
	l := 1024
	buf := concatByteSlices(
		newUint64SliceEncoded(l),
	)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder := NewBorshDecoder(buf)

		got := make([]uint64, l)
		rv := reflect.ValueOf(&got).Elem()
		k := rv.Type().Elem().Kind()

		err := reflect_readArrayOfUint_(decoder, len(buf)/8, k, rv, LE)
		if err != nil {
			b.Error(err)
		}
		if len(got) != l {
			b.Errorf("got %d, want %d", len(got), l)
		}
	}
}
