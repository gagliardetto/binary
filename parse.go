package bin

import (
	"encoding/binary"
	"reflect"
	"strconv"
	"strings"
)

type fieldTag struct {
	SizeOf                string
	SliceOffsetOf         string
	SliceOffsetMultiplier int
	Skip                  bool
	Order                 binary.ByteOrder
	Optional              bool
	BinaryExtension       bool
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
		} else if strings.HasPrefix(s, "sliceoffsetof=") {
			tmp := strings.SplitN(s, "=", 2)
			tmp = strings.SplitN(tmp[1], ",", 2)
			t.SliceOffsetOf = tmp[0]
			multiplier, err := strconv.Atoi(tmp[1])
			if err != nil {
				panic("slice offset multiplier must be a valid int, got: " + tmp[1])
			}
			t.SliceOffsetMultiplier = multiplier
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
	return t
}
