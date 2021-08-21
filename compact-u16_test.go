package bin

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompactU16(t *testing.T) {
	candidates := []int{3, 0x7f, 0x7f + 1, 0x3fff, 0x3fff + 1}
	for _, val := range candidates {
		buf := make([]byte, 0)
		EncodeCompactU16Length(&buf, val)

		buf = append(buf, []byte("hello world")...)
		decoded := DecodeCompactU16Length(buf)

		require.Equal(t, val, decoded)
	}
	for _, val := range candidates {
		buf := make([]byte, 0)
		EncodeCompactU16Length(&buf, val)

		buf = append(buf, []byte("hello world")...)
		decoded, err := DecodeCompactU16LengthFromByteReader(bytes.NewReader(buf))
		if err != nil {
			panic(err)
		}

		require.Equal(t, val, decoded)
	}
}
