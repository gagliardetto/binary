package bin

import (
	"encoding/binary"
	"reflect"
	"strings"
)

type filedTag struct {
	Sizeof   string
	Skip     bool
	Order    binary.ByteOrder

}


func parseFieldTag(tag reflect.StructTag) *filedTag {
	t := &filedTag{
		Order: binary.LittleEndian,
	}
	if tag == "-" {
		t.Skip = true
		return t
	}
	tagStr := tag.Get("bin")
	for _, s := range strings.Split(tagStr, ",") {
		if strings.HasPrefix(s, "sizeof=") {
			tmp := strings.SplitN(s, "=", 2)
			t.Sizeof = tmp[1]
		} else if s == "big" {
			t.Order = binary.BigEndian
		} else if s == "little" {
			t.Order = binary.LittleEndian
		}
	}
	return t
}