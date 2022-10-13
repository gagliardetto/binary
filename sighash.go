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
	"crypto/sha256"
	"strings"
	"unicode"
)

// Sighash creates an anchor sighash for the provided namespace and element.
// An anchor sighash is the first 8 bytes of the sha256 of {namespace}:{name}
// NOTE: you must first convert the name to snake case using `ToSnakeForSighash`.
func Sighash(namespace string, name string) []byte {
	data := namespace + ":" + name
	sum := sha256.Sum256([]byte(data))
	return sum[0:8]
}

func SighashInstruction(name string) []byte {
	return Sighash(SIGHASH_GLOBAL_NAMESPACE, ToSnakeForSighash(name))
}

func SighashAccount(name string) []byte {
	return Sighash(SIGHASH_ACCOUNT_NAMESPACE, ToSnakeForSighash(name))
}

func SighashTypeID(namespace string, name string) TypeID {
	return TypeIDFromBytes(Sighash(namespace, ToSnakeForSighash(name)))
}

// Namespace for calculating state instruction sighash signatures.
const SIGHASH_STATE_NAMESPACE string = "state"

// Namespace for calculating instruction sighash signatures for any instruction
// not affecting program state.
const SIGHASH_GLOBAL_NAMESPACE string = "global"

const SIGHASH_ACCOUNT_NAMESPACE string = "account"

const ACCOUNT_DISCRIMINATOR_SIZE = 8

// https://github.com/project-serum/anchor/pull/64/files
// https://github.com/project-serum/anchor/blob/2f780e0d274f47e442b3f0d107db805a41c6def0/ts/src/coder/common.ts#L109
// https://github.com/project-serum/anchor/blob/6b5ed789fc856408986e8868229887354d6d4073/lang/syn/src/codegen/program/common.rs#L17

// TODO:
// https://github.com/project-serum/anchor/blob/84a2b8200cc3c7cb51d7127918e6cbbd836f0e99/ts/src/error.ts#L48

func ToSnakeForSighash(s string) string {
	return ToRustSnakeCase(s)
}

type reader struct {
	runes []rune
	index int
}

func splitStringByRune(s string) []rune {
	var res []rune
	iterateStringAsRunes(s, func(r rune) bool {
		res = append(res, r)
		return true
	})
	return res
}

func iterateStringAsRunes(s string, callback func(r rune) bool) {
	for _, rn := range s {
		doContinue := callback(rn)
		if !doContinue {
			return
		}
	}
}

func newReader(s string) *reader {
	return &reader{
		runes: splitStringByRune(s),
		index: -1,
	}
}

func (r reader) This() (int, rune) {
	return r.index, r.runes[r.index]
}

func (r reader) HasNext() bool {
	return r.index < len(r.runes)-1
}

func (r reader) Peek() (int, rune) {
	if r.HasNext() {
		return r.index + 1, r.runes[r.index+1]
	}
	return -1, rune(0)
}

func (r *reader) Move() bool {
	if r.HasNext() {
		r.index++
		return true
	}
	return false
}

// #[cfg(not(feature = "unicode"))]
func splitIntoWords(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	return parts
}

type _WordMode int

const (
	/// There have been no lowercase or uppercase characters in the current
	/// word.
	_Boundary _WordMode = iota
	/// The previous cased character in the current word is lowercase.
	_Lowercase
	/// The previous cased character in the current word is uppercase.
	_Uppercase
)

// ToRustSnakeCase converts the given string to a snake_case string.
// Ported from https://github.com/withoutboats/heck/blob/c501fc95db91ce20eaef248a511caec7142208b4/src/lib.rs#L75 as used by Anchor.
func ToRustSnakeCase(s string) string {
	builder := new(strings.Builder)

	first_word := true
	words := splitIntoWords(s)
	for _, word := range words {
		char_indices := newReader(word)
		init := 0
		mode := _Boundary

		for char_indices.Move() {
			i, c := char_indices.This()

			// Skip underscore characters
			if c == '_' {
				if init == i {
					init += 1
				}
				continue
			}

			if next_i, next := char_indices.Peek(); next_i != -1 {

				// The mode including the current character, assuming the
				// current character does not result in a word boundary.
				next_mode := func() _WordMode {
					if unicode.IsLower(c) {
						return _Lowercase
					} else if unicode.IsUpper(c) {
						return _Uppercase
					} else {
						return mode
					}
				}()

				// Word boundary after if next is underscore or current is
				// not uppercase and next is uppercase
				if next == '_' || (next_mode == _Lowercase && unicode.IsUpper(next)) {
					if !first_word {
						// boundary(f)?;
						builder.WriteRune('_')
					}
					{
						// with_word(&word[init..next_i], f)?;
						builder.WriteString(strings.ToLower(word[init:next_i]))
					}

					first_word = false
					init = next_i
					mode = _Boundary

					// Otherwise if current and previous are uppercase and next
					// is lowercase, word boundary before
				} else if mode == _Uppercase && unicode.IsUpper(c) && unicode.IsLower(next) {
					if !first_word {
						// boundary(f)?;
						builder.WriteRune('_')
					} else {
						first_word = false
					}
					{
						// with_word(&word[init..i], f)?;
						builder.WriteString(strings.ToLower(word[init:i]))
					}
					init = i
					mode = _Boundary

					// Otherwise no word boundary, just update the mode
				} else {
					mode = next_mode
				}

			} else {
				// Collect trailing characters as a word
				if !first_word {
					// boundary(f)?;
					builder.WriteRune('_')
				} else {
					first_word = false
				}
				{
					// with_word(&word[init..], f)?;
					builder.WriteString(strings.ToLower(word[init:]))
				}
				break
			}
		}
	}

	return builder.String()
}
