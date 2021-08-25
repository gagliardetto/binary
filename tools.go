package bin

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatByteSlice formats the given byte slice into a readable format.
func FormatByteSlice(buf []byte) string {
	elems := make([]string, 0)
	for _, v := range buf {
		elems = append(elems, strconv.Itoa(int(v)))
	}

	return "[" + strings.Join(elems, ", ") + "]" + fmt.Sprintf("(len=%v)", len(elems))
}

type WriteByWrite struct {
	writes [][]byte
	name   string
}

func NewWriteByWrite(name string) *WriteByWrite {
	return &WriteByWrite{
		name: name,
	}
}

func (rec *WriteByWrite) Write(b []byte) (int, error) {
	rec.writes = append(rec.writes, b)
	return len(b), nil
}

func (rec *WriteByWrite) Bytes() []byte {
	out := make([]byte, 0)
	for _, v := range rec.writes {
		out = append(out, v...)
	}
	return out
}

func (rec WriteByWrite) String() string {
	builder := new(strings.Builder)
	if rec.name != "" {
		builder.WriteString(rec.name + ":\n")
	}
	for index, v := range rec.writes {
		builder.WriteString(fmt.Sprintf("- %v: %s\n", index, FormatByteSlice(v)))
	}
	return builder.String()
}

// IsByteSlice returns true if the provided element is a []byte.
func IsByteSlice(v interface{}) bool {
	_, ok := v.([]byte)
	return ok
}
