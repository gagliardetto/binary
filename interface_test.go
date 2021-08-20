package bin

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Example struct {
	Prefix byte
	Value  uint32
}

func (e *Example) UnmarshalBinary(decoder *Decoder) (err error) {
	if e.Prefix, err = decoder.ReadByte(); err != nil {
		return err
	}
	if e.Value, err = decoder.ReadUint32(BE()); err != nil {
		return err
	}
	return nil
}

func (e *Example) MarshalBinary(encoder *Encoder) error {
	if err := encoder.WriteByte(e.Prefix); err != nil {
		return err
	}
	return encoder.WriteUint32(e.Value, BE())
}

func TestMarshalBinary(t *testing.T) {
	buf := new(bytes.Buffer)
	e := &Example{Value: 72, Prefix: 0xaa}
	enc := NewBinEncoder(buf)
	enc.Encode(e)

	assert.Equal(t, []byte{
		0xaa, 0x00, 0x00, 0x00, 0x48,
	}, buf.Bytes())
}

func TestUnmarshalBinary(t *testing.T) {
	buf := []byte{
		0xaa, 0x00, 0x00, 0x00, 0x48,
	}

	e := &Example{}
	d := NewBinDecoder(buf)
	err := d.Decode(e)
	assert.NoError(t, err)
	assert.Equal(t, e, &Example{Value: 72, Prefix: 0xaa})
	assert.Equal(t, 0, d.Remaining())
}
