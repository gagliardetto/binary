package bin

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseFieldTag(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectValue *fieldTag
	}{
		{
			name: "no tags",
			tag:  "",
			expectValue: &fieldTag{
				Order: binary.LittleEndian,
			},
		},
		{
			name: "with a skip",
			tag:  `bin:"-"`,
			expectValue: &fieldTag{
				Order: binary.LittleEndian,
				Skip:  true,
			},
		},
		{
			name: "with a sizeof",
			tag:  `bin:"sizeof=Tokens"`,
			expectValue: &fieldTag{
				Order:  binary.LittleEndian,
				Sizeof: "Tokens",
			},
		},
		{
			name: "with a optional",
			tag:  `bin:"optional"`,
			expectValue: &fieldTag{
				Order:    binary.LittleEndian,
				Optional: true,
			},
		},
		{
			name: "with a optional and size of",
			tag:  `bin:"optional,sizeof=Nodes"`,
			expectValue: &fieldTag{
				Order:    binary.LittleEndian,
				Optional: true,
				Sizeof:   "Nodes",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectValue, parseFieldTag(reflect.StructTag(test.tag)))
		})
	}

}
