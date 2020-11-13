# dfuse binary
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/dfuse-io/logging)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This is the binary library used as part of **[dfuse](https://github.com/dfuse-io/dfuse)**.

Usage
----

#### Struct Field Tags
- `_` will skip the field when encoding & deocode a struc
- `sizeof=` indicates this field is a number used to track the length of a another field.
- Bare values will be parsed as type and little endian when necessary

#### Supported Types
 - `int8`, `int16`, `int32`, `int64`, `Int128`
 - `uint8`, `uint16`,`uint32`,`uint64`, `Uint128`
 - `float32`, `float64`, `Float128`
 - `string`, `bool`
 - `Varint16`, `Varint32`
 - `Varuint16`, `Varuint32`
#### Custom Types
To implement custom types, your types would need to implement the `MarshalerBinary` & `UnmarshalerBinary` interfaces

Example
----
```Go
type Example struct {
    Var      uint32 `bin:"_"`
    Str      string
    IntCount utin32 `bin:"sizeof=Var"`
    Weird    [8]byte
    Var      []int
}
```
