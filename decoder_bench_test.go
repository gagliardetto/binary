package bin

import (
	"testing"
)

func BenchmarkDecodeUintSlice(b *testing.B) {
	buf := concatByteSlices(
		// length:
		uint32ToBytes(8, LE),
		// data:
		uint32ToBytes(0, LE),
		uint32ToBytes(1, LE),
		uint32ToBytes(2, LE),
		uint32ToBytes(3, LE),
		uint32ToBytes(4, LE),
		uint32ToBytes(5, LE),
		uint32ToBytes(6, LE),
		uint32ToBytes(7, LE),
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
