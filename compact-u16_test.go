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
		EncodeLength(&buf, val)

		buf = append(buf, []byte("hello world")...)
		decoded := DecodeLength(buf)

		require.Equal(t, val, decoded)
	}
	for _, val := range candidates {
		buf := make([]byte, 0)
		EncodeLength(&buf, val)

		buf = append(buf, []byte("hello world")...)
		decoded, err := DecodeLengthFromByteReader(bytes.NewReader(buf))
		if err != nil {
			panic(err)
		}

		require.Equal(t, val, decoded)
	}
}
