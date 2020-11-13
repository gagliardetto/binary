package bin

import "github.com/dfuse-io/logging"

func init() {
	logging.TestingOverride()
	//traceEnabled = true
	//zlog, _ = zap.NewDevelopment()
}

type aliasTestType uint64

type unknownType struct {
}

type binaryInvalidTestStruct struct {
	F1 unknownType
}

type binaryTestStruct struct {
	F1  string
	F2  int16
	F3  uint16
	F4  int32
	F5  uint32
	F6  int64
	F7  uint64
	F8  float32
	F9  float64
	F10 []string
	F11 [2]string
	F12 byte
	F13 []byte
	F14 bool
	F15 Int64
	F16 Uint64
	F17 JSONFloat64
	F18 Uint128
	F19 Int128
	F20 Float128
	F21 Varuint32
	F22 Varint32
	F23 Bool
	F24 HexBytes
}
