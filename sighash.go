package bin

import (
	"crypto/sha256"

	"github.com/davecgh/go-spew/spew"
	. "github.com/gagliardetto/utilz"
)

func Sighash(namespace string, name string) []byte {
	methodName := name

	data := namespace + ":" + methodName

	sum := sha256.Sum256([]byte(data))
	spew.Dump(namespace, name, data, sum)
	Ln(FormatByteSlice(sum[:]))

	return sum[0:8]
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
func InstructionSigHash(name string) []byte {
	return Sighash(SIGHASH_GLOBAL_NAMESPACE, ToLowerCamel(name))
}

// TODO:
// https://github.com/project-serum/anchor/blob/84a2b8200cc3c7cb51d7127918e6cbbd836f0e99/ts/src/error.ts#L48
