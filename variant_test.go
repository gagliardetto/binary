package bin

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Forest struct {
	T Tree
}

type Tree struct {
	Padding   [5]byte
	NodeCount uint32 `bin:"sizeof=Nodes"`
	Random    uint64
	Nodes     []*Node
}

var NodeVariantDef = NewVariantDefinition(Uint32TypeIDEncoding, []VariantType{
	{"left_node", (*NodeLeft)(nil)},
	{"right_node", (*NodeRight)(nil)},
	{"inner_node", (*NodeInner)(nil)},
})

type Node struct {
	BaseVariant
}

type NodeLeft struct {
	Key         uint32
	Description string
}

type NodeRight struct {
	Owner    uint64
	Padding  [2]byte
	Quantity Uint64
}

type NodeInner struct {
	Key Uint128
}

func (n *Node) UnmarshalBinary(decoder *Decoder) error {
	return n.BaseVariant.UnmarshalBinaryVariant(decoder, NodeVariantDef)
}
func (n *Node) MarshalBinary(encoder *Encoder) error {
	err := encoder.WriteUint32(n.TypeID)
	if err != nil {
		return err
	}
	return encoder.Encode(n.Impl)
}

func TestDecoder_AnotherTest(t *testing.T) {
	buf := []byte{
		0x4f, 0x9f, 0x3, 0x00, // 237391
	}

	d := NewDecoder(buf)
	var v uint32
	err := d.Decode(&v)
	require.NoError(t, err)
	fmt.Println("value ", v)
}

func TestDecode_Variant(t *testing.T) {
	buf := []byte{
		0x73, 0x65, 0x72, 0x75, 0x6d, // Padding[5]byte
		0x05, 0x00, 0x00, 0x00, // Node length 5
		0xff, 0xff, 0x00, 0x00, 0x00, 0x0, 0x00, 0x00, // ROOT  65,535
		0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x03, 0x61, 0x62, 0x63, // left node -> key = 3, description "abc"
		0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // right node -> owner = 3, quantity 13
		0x01, 0x00, 0x00, 0x00, 0x52, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x9b, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // right node -> owner = 82, quantity 923
		0x02, 0x00, 0x00, 0x00, 0xff, 0x7f, 0xc6, 0xa4, 0x7e, 0x8d, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // inner node -> key = 999999999999999
		0x02, 0x00, 0x00, 0x00, 0x23, 0xd3, 0xd8, 0x9a, 0x99, 0x7e, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // inner node -> key = 983623123129123
	}

	decoder := NewDecoder(buf)
	forest := Forest{}
	err := decoder.Decode(&forest)
	require.NoError(t, err)
	require.Equal(t, 0, decoder.remaining())
	assert.Equal(t, Tree{
		Padding:   [5]byte{0x73, 0x65, 0x72, 0x75, 0x6d},
		NodeCount: 5,
		Random:    65535,
		Nodes: []*Node{
			{
				BaseVariant: BaseVariant{
					TypeID: 0,
					Impl: &NodeLeft{
						Key:         3,
						Description: "abc",
					},
				},
			},
			{
				BaseVariant: BaseVariant{
					TypeID: 1,
					Impl: &NodeRight{
						Owner:    3,
						Padding:  [2]byte{0x00, 0x00},
						Quantity: 13,
					},
				},
			},
			{
				BaseVariant: BaseVariant{
					TypeID: 1,
					Impl: &NodeRight{
						Owner:    82,
						Padding:  [2]byte{0x00, 0x00},
						Quantity: 923,
					},
				},
			},
			{
				BaseVariant: BaseVariant{
					TypeID: 2,
					Impl: &NodeInner{
						Key: Uint128{
							Lo: 999999999999999,
							Hi: 0,
						},
					},
				},
			},
			{
				BaseVariant: BaseVariant{
					TypeID: 2,
					Impl: &NodeInner{
						Key: Uint128{
							Lo: 983623123129123,
							Hi: 0,
						},
					},
				},
			},
		},
	}, forest.T)

}

type unexportesStruct struct {
	value uint32
}

func Test_Decoding_UnexporterStruct(t *testing.T) {
	buf := []byte{
		0x05, 0x00, 0x00, 0x00,
	}

	decoder := NewDecoder(buf)
	s := unexportesStruct{}
	err := decoder.Decode(&s)
	require.NoError(t, err)
	require.Equal(t, 4, decoder.remaining())
	assert.Equal(t, unexportesStruct{value: 0}, s)
}
