package bin

import (
	"encoding/binary"
	"reflect"
	"strings"
)

type fieldTag struct {
	SizeOf          string
	Skip            bool
	Order           binary.ByteOrder
	Optional        bool
	BinaryExtension bool

	IsBorshEnum bool
}

func parseFieldTag(tag reflect.StructTag) *fieldTag {
	t := &fieldTag{
		Order: binary.LittleEndian,
	}
	tagStr := tag.Get("bin")
	for _, s := range strings.Split(tagStr, " ") {
		if strings.HasPrefix(s, "sizeof=") {
			tmp := strings.SplitN(s, "=", 2)
			t.SizeOf = tmp[1]
		} else if s == "big" {
			t.Order = binary.BigEndian
		} else if s == "little" {
			t.Order = binary.LittleEndian
		} else if s == "optional" {
			t.Optional = true
		} else if s == "binary_extension" {
			t.BinaryExtension = true
		} else if s == "-" {
			t.Skip = true
		}
	}

	// TODO: parse other borsh tags
	if strings.TrimSpace(tag.Get("borsh_skip")) == "true" {
		t.Skip = true
	}
	if strings.TrimSpace(tag.Get("borsh_enum")) == "true" {
		t.IsBorshEnum = true
	}
	return t
}
