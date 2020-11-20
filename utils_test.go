package bin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_twosComplement(t *testing.T) {
	tests := []struct {
		name   string
		in     []byte
		expect []byte
	}{
		{
			name:   "empty array",
			in:     []byte{},
			expect: []byte{0x1},
		},
		{
			name:   "one element",
			in:     []byte{0x01},
			expect: []byte{0xff},
		},
		{
			name:   "basic test",
			in:     []byte{0xaa, 0xbb, 0xcc},
			expect: []byte{0x55, 0x44, 0x34},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, twosComplement(test.in))
		})
	}

}
