// Copyright 2021 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bin

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompactU16(t *testing.T) {
	candidates := []int{3, 0x7f, 0x7f + 1, 0x3fff, 0x3fff + 1}
	for _, val := range candidates {
		buf := make([]byte, 0)
		EncodeCompactU16Length(&buf, val)

		buf = append(buf, []byte("hello world")...)
		decoded := DecodeCompactU16Length(buf)

		require.Equal(t, val, decoded)
	}
	for _, val := range candidates {
		buf := make([]byte, 0)
		EncodeCompactU16Length(&buf, val)

		buf = append(buf, []byte("hello world")...)
		{
			decoded, err := DecodeCompactU16LengthFromByteReader(bytes.NewReader(buf))
			if err != nil {
				panic(err)
			}
			require.Equal(t, val, decoded)
		}
		{
			decoded, _, err := DecodeCompactU16(buf)
			if err != nil {
				panic(err)
			}
			require.Equal(t, val, decoded)
		}
	}
}

func BenchmarkCompactU16(b *testing.B) {
	// generate 1000 random values
	candidates := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		candidates[i] = i
	}

	buf := make([]byte, 0)
	EncodeCompactU16Length(&buf, math.MaxUint16)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = DecodeCompactU16(buf)
	}
}

func BenchmarkCompactU16Reader(b *testing.B) {
	// generate 1000 random values
	candidates := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		candidates[i] = i
	}

	buf := make([]byte, 0)
	EncodeCompactU16Length(&buf, math.MaxUint16)

	reader := NewBorshDecoder(buf)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reader.ReadCompactU16()
		reader.SetPosition(0)
	}
}
