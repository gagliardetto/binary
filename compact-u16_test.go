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
		decoded, err := DecodeCompactU16LengthFromByteReader(bytes.NewReader(buf))
		if err != nil {
			panic(err)
		}

		require.Equal(t, val, decoded)
	}
}
