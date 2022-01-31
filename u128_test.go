package bin

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestUint128(t *testing.T) {
	// from bytes:
	data := []byte{51, 47, 223, 255, 255, 255, 255, 255, 30, 12, 0, 0, 0, 0, 0, 0}

	orderID, err := decimal.NewFromString("57240246860720736513843")
	if err != nil {
		panic(err)
	}
	spew.Dump(orderID)

	{
		u128 := NewUint128LittleEndian()
		err := u128.UnmarshalWithDecoder(NewBorshDecoder(data))
		require.NoError(t, err)
		require.Equal(t, uint64(3102), u128.Hi)
		require.Equal(t, uint64(18446744073707401011), u128.Lo)
		require.Equal(t, orderID.BigInt(), u128.BigInt())
		require.Equal(t, orderID.String(), u128.DecimalString())
	}
	{
		u128 := NewUint128BigEndian()
		ReverseBytes(data)
		err := u128.UnmarshalWithDecoder(NewBorshDecoder(data))
		require.NoError(t, err)
		require.Equal(t, uint64(3102), u128.Hi)
		require.Equal(t, uint64(18446744073707401011), u128.Lo)
		require.Equal(t, orderID.BigInt(), u128.BigInt())
		require.Equal(t, orderID.String(), u128.DecimalString())
	}
	{
		j := []byte(`{"i":"57240246860720736513843"}`)
		var object struct {
			I Uint128 `json:"i"`
		}

		err := json.Unmarshal(j, &object)
		require.NoError(t, err)
		require.Equal(t, uint64(3102), object.I.Hi)
		require.Equal(t, uint64(18446744073707401011), object.I.Lo)
		require.Equal(t, orderID.BigInt(), object.I.BigInt())
		require.Equal(t, orderID.String(), object.I.DecimalString())

		{
			out, err := json.Marshal(object)
			require.NoError(t, err)
			require.Equal(t, j, out)
		}
	}
}
