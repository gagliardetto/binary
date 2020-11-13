package bin

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_parseFieldTag(t *testing.T) {
	tests := []struct{
		name string
		tag string
		expectValue *filedTag
	}{
		{
			name: "no tags",
			tag: "",
			expectValue: &filedTag{
				Order:  binary.LittleEndian,
			},
		},
		{
			name: "with a skip",
			tag: "-",
			expectValue: &filedTag{
				Order:  binary.LittleEndian,
				Skip: true,
			},
		},
		{
			name: "with a sizeof",
			tag: `bin:"sizeof=Tokens"`,
			expectValue: &filedTag{
				Order:  binary.LittleEndian,
				Sizeof:  "Tokens",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectValue, parseFieldTag(reflect.StructTag(test.tag)))
		})
	}

}