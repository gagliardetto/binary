package bin

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)


type SlabUninitialized struct {
	Padding [68]byte `json:"-"`
}

type SlabInnerNode struct {
	//    u32('prefixLen'),
	//    u128('key'),
	//    seq(u32(), 2, 'children'),
	PrefixLen uint32
	Key       Uint128
	Children  [2]uint32
	Padding [40]byte `json:"-"`
}


type SlabLeafNode struct {
	OwnerSlot     uint8
	FeeTier       uint8
	Padding       [2]byte `json:"-"`
	Key           Uint128
	Owner         PublicKey
	Quantity      Uint64
	ClientOrderId Uint64
}
type SlabFreeNode struct {
	Next uint32
	Padding [64]byte `json:"-"`
}

type SlabLastFreeNode struct {
	Padding [68]byte `json:"-"`
}

type PublicKey [32]byte

var SlabFactoryImplDef = NewVariantDefinition(Uint32TypeIDEncoding, []VariantType{
	{"uninitialized", (*SlabUninitialized)(nil)},
	{"inner_node", (*SlabInnerNode)(nil)},
	{"leaf_node", (*SlabLeafNode)(nil)},
	{"free_node", (*SlabFreeNode)(nil)},
	{"last_free_node", (*SlabLastFreeNode)(nil)},
})

type Slab struct {
	BaseVariant
}

func (s *Slab) UnmarshalBinary(decoder *Decoder) error {
	return s.BaseVariant.UnmarshalBinaryVariant(decoder, SlabFactoryImplDef)
}
func (s *Slab) MarshalBinary(encoder *Encoder) error {
	err := encoder.writeUint32(s.TypeID)
	if err != nil {
		return err
	}
	return encoder.Encode(s.Impl)
}

type Orderbook struct {
	// ORDERBOOK_LAYOUT
	SerumPadding [5]byte `json:"-"`
	AccountFlags uint64
	// SLAB_LAYOUT
	// SLAB_HEADER_LAYOUT
	BumpIndex    uint32  `bin:"sizeof=Nodes"`
	ZeroPaddingA [4]byte `json:"-"`
	FreeListLen  uint32
	ZeroPaddingB [4]byte `json:"-"`
	FreeListHead uint32
	Root         uint32
	LeafCount    uint32
	ZeroPaddingC [4]byte `json:"-"`

	// SLAB_NODE_LAYOUT
	Nodes []*Slab `bin: ""`
}


func TestDecoder_DecodeSol(t *testing.T) {
	hexData, err := ioutil.ReadFile("./testdata/orderbook_lite.hex")
	require.NoError(t, err)

	fmt.Println(hexData)

	cnt, err := hex.DecodeString(string(hexData))
	require.NoError(t, err)

	decoder := NewDecoder(cnt)
	var ob *Orderbook
	err = decoder.Decode(&ob)
	require.NoError(t, err)

	require.Equal(t, 0, decoder.remaining())
	json, err := json.MarshalIndent(ob, "", "   ")
	require.NoError(t, err)
	fmt.Println(string(json))

	fmt.Println("-------------------------------------------------------")
	fmt.Println("-------------------------------------------------------")
	fmt.Println("-------------------------------------------------------")

	buf := new(bytes.Buffer)
	encoder := NewEncoder(buf)
	err = encoder.Encode(ob)
	require.NoError(t, err)

	obHex := hex.EncodeToString(buf.Bytes())

	fmt.Println("expected:", hexData)
	fmt.Println("actual  :", obHex)
	require.Equal(t, cnt, buf.Bytes())
}


func TestDecoder_Slabs(t *testing.T) {

	//zlog, _ := zap.NewDevelopment()
	//EnableDebugLogging(zlog)

	rawSlabs := []string{
		"0100000035000000010babffffffffff4105000000000000400000003f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"0200000014060000b2cea5ffffffffff23070000000000005ae01b52d00a090c6dc6fce8e37a225815cff2223a99c6dfdad5aae56d3db670e62c000000000000140b0fadcf8fcebf",
		"030000003400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	}

	for _, s := range rawSlabs {
		cnt, err := hex.DecodeString(s)
		require.NoError(t, err)

		decoder := NewDecoder(cnt)
		var slab *Slab
		err = decoder.Decode(&slab)
		require.NoError(t, err)

		json, err := json.MarshalIndent(slab, "", "   ")
		require.NoError(t, err)
		fmt.Println(string(json))

		//require.Equal(t, 0, decoder.remaining())

	}
}
