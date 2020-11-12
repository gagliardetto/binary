package bin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/tidwall/gjson"
)

//
/// Variant (emulates `fc::static_variant` type)
//

type Variant interface {
	Assign(typeID uint, impl interface{})
	Obtain() (typeID uint, impl interface{})
}

type VariantType struct {
	Name string
	Type interface{}
}

type VariantDefinition struct {
	typeIDToType   map[uint32]reflect.Type
	typeIDToName   map[uint32]string
	typeNameToID   map[string]uint32
	typeIDEncoding TypeIDEncoding
}

type TypeIDEncoding uint32

const (
	Uvarint32TypeIDEncoding TypeIDEncoding = iota
	Uint32TypeIDEncoding
)

// NewVariantDefinition creates a variant definition based on the *ordered* provided types.
// It's the ordering that defines the binary variant value just like in native `nodeos` C++
// and in Smart Contract via the `std::variant` type. It's important to pass the entries
// in the right order!
//
// This variant definition can now be passed to functions of `BaseVariant` to implement
// marshal/unmarshaling functionalities for binary & JSON.
func NewVariantDefinition(typeIDEncoding TypeIDEncoding, types []VariantType) (out *VariantDefinition) {
	if len(types) < 0 {
		panic("it's not valid to create a variant definition without any types")
	}

	typeCount := len(types)
	out = &VariantDefinition{
		typeIDEncoding: typeIDEncoding,
		typeIDToType:   make(map[uint32]reflect.Type, typeCount),
		typeIDToName:   make(map[uint32]string, typeCount),
		typeNameToID:   make(map[string]uint32, typeCount),
	}

	for i, typeDef := range types {
		typeID := uint32(i)

		// FIXME: Check how the reflect.Type is used and cache all its usage in the definition.
		//        Right now, on each Unmarshal, we re-compute some expensive stuff that can be
		//        re-used like the `typeGo.Elem()` which is always the same. It would be preferable
		//        to have those already pre-defined here so we can actually speed up the
		//        Unmarshal code.
		out.typeIDToType[typeID] = reflect.TypeOf(typeDef.Type)
		out.typeIDToName[typeID] = typeDef.Name
		out.typeNameToID[typeDef.Name] = typeID
	}

	return out
}

func (d *VariantDefinition) TypeID(name string) uint32 {
	id, found := d.typeNameToID[name]
	if !found {
		knownNames := make([]string, len(d.typeNameToID))
		i := 0
		for name := range d.typeNameToID {
			knownNames[i] = name
			i++
		}

		panic(fmt.Errorf("trying to use an unknown type name %q, known names are %q", name, strings.Join(knownNames, ", ")))
	}

	return id
}

type VariantImplFactory = func() interface{}
type OnVariant = func(impl interface{}) error

type BaseVariant struct {
	TypeID uint32
	Impl   interface{}
}

func (a *BaseVariant) Assign(typeID uint32, impl interface{}) {
	a.TypeID = typeID
	a.Impl = impl
}

func (a *BaseVariant) Obtain(def *VariantDefinition) (typeID uint32, typeName string, impl interface{}) {
	return a.TypeID, def.typeIDToName[a.TypeID], a.Impl
}

func (a *BaseVariant) MarshalJSON(def *VariantDefinition) ([]byte, error) {
	typeName, found := def.typeIDToName[a.TypeID]
	if !found {
		return nil, fmt.Errorf("type %d is not know by variant definition", a.TypeID)
	}

	return json.Marshal([]interface{}{typeName, a.Impl})
}

func (a *BaseVariant) UnmarshalJSON(data []byte, def *VariantDefinition) error {
	typeResult := gjson.GetBytes(data, "0")
	implResult := gjson.GetBytes(data, "1")

	if !typeResult.Exists() || !implResult.Exists() {
		return fmt.Errorf("invalid format, expected '[<type>, <impl>]' pair, got %q", string(data))
	}

	typeName := typeResult.String()
	typeID, found := def.typeNameToID[typeName]
	if !found {
		return fmt.Errorf("type %q is not know by variant definition", typeName)
	}

	typeGo := def.typeIDToType[typeID]
	if typeGo == nil {
		return fmt.Errorf("no known type for %q", typeName)
	}

	a.TypeID = typeID

	if typeGo.Kind() == reflect.Ptr {
		a.Impl = reflect.New(typeGo.Elem()).Interface()
		if err := json.Unmarshal([]byte(implResult.Raw), a.Impl); err != nil {
			return err
		}
	} else {
		// This is not the most optimal way of doing things for "value"
		// types (over "pointer" types) as we always allocate a new pointer
		// element, unmarshal it and then either keep the pointer type or turn
		// it into a value type.
		//
		// However, in non-reflection based code, one would do like this and
		// avoid an `new` memory allocation:
		//
		// ```
		// name := eos.Name("")
		// json.Unmarshal(data, &name)
		// ```
		//
		// This would work without a problem. In reflection code however, I
		// did not find how one can go from `reflect.Zero(typeGo)` (which is
		// the equivalence of doing `name := eos.Name("")`) and take the
		// pointer to it so it can be unmarshalled correctly.
		//
		// A played with various iteration, and nothing got it working. Maybe
		// the next step would be to explore the `unsafe` package and obtain
		// an unsafe pointer and play with it.
		value := reflect.New(typeGo)
		if err := json.Unmarshal([]byte(implResult.Raw), value.Interface()); err != nil {
			return err
		}

		a.Impl = value.Elem().Interface()
	}

	return nil
}

func ptr(v reflect.Value) reflect.Value {
	pt := reflect.PtrTo(v.Type())
	pv := reflect.New(pt.Elem())
	pv.Elem().Set(v)
	return pv
}

func (a *BaseVariant) UnmarshalBinaryVariant(decoder *Decoder, def *VariantDefinition) (err error) {

	var typeID uint32
	switch def.typeIDEncoding {
	case Uvarint32TypeIDEncoding:
		typeID, err = decoder.ReadUvarint32()
		if err != nil {
			return fmt.Errorf("uvarint32: unable to read variant type id: %s", err)
		}
	case Uint32TypeIDEncoding:
		typeID, err = decoder.ReadUint32()
		if err != nil {
			return fmt.Errorf("uint32: unable to read variant type id: %s", err)
		}
	}

	a.TypeID = typeID
	typeGo := def.typeIDToType[typeID]
	if typeGo == nil {
		return fmt.Errorf("no known type for type %d", typeID)
	}

	if typeGo.Kind() == reflect.Ptr {
		a.Impl = reflect.New(typeGo.Elem()).Interface()
		if err = decoder.Decode(a.Impl); err != nil {
			return fmt.Errorf("unable to decode variant type %d: %s", typeID, err)
		}
	} else {
		// This is not the most optimal way of doing things for "value"
		// types (over "pointer" types) as we always allocate a new pointer
		// element, unmarshal it and then either keep the pointer type or turn
		// it into a value type.
		//
		// However, in non-reflection based code, one would do like this and
		// avoid an `new` memory allocation:
		//
		// ```
		// name := eos.Name("")
		// json.Unmarshal(data, &name)
		// ```
		//
		// This would work without a problem. In reflection code however, I
		// did not find how one can go from `reflect.Zero(typeGo)` (which is
		// the equivalence of doing `name := eos.Name("")`) and take the
		// pointer to it so it can be unmarshalled correctly.
		//
		// A played with various iteration, and nothing got it working. Maybe
		// the next step would be to explore the `unsafe` package and obtain
		// an unsafe pointer and play with it.
		value := reflect.New(typeGo)
		if err = decoder.Decode(value.Interface()); err != nil {
			return fmt.Errorf("unable to decode variant type %d: %s", typeID, err)
		}

		a.Impl = value.Elem().Interface()
	}
	return nil
}

// Implementation of `fc::variant` types

type fcVariantType uint32

const (
	fcVariantNullType fcVariantType = iota
	fcVariantInt64Type
	fcVariantUint64Type
	fcVariantDoubleType
	fcVariantBoolType
	fcVariantStringType
	fcVariantArrayType
	fcVariantObjectType
	fcVariantBlobType
)

func (t fcVariantType) String() string {
	switch t {
	case fcVariantNullType:
		return "null"
	case fcVariantInt64Type:
		return "int64"
	case fcVariantUint64Type:
		return "uint64"
	case fcVariantDoubleType:
		return "double"
	case fcVariantBoolType:
		return "bool"
	case fcVariantStringType:
		return "string"
	case fcVariantArrayType:
		return "array"
	case fcVariantObjectType:
		return "object"
	case fcVariantBlobType:
		return "blob"
	}

	return "unknown"
}

// FIXME: Ideally, we would re-use `BaseVariant` but that requires some
//        re-thinking of the decoder to make it efficient to read FCVariant types. For now,
//        let's re-code it a bit to make it as efficient as possible.
type fcVariant struct {
	TypeID fcVariantType
	Impl   interface{}
}

func (a fcVariant) IsNil() bool {
	return a.TypeID == fcVariantNullType
}

// ToNative transform the actual implementation, walking each sub-element like array
// and object, turning everything along the way in Go primitives types.
//
// **Note** For `Int64` and `Uint64`, we return `eos.Int64` and `eos.Uint64` types
//          so that JSON marshalling is done correctly for large numbers
func (a fcVariant) ToNative() interface{} {
	if a.TypeID == fcVariantNullType ||
		a.TypeID == fcVariantDoubleType ||
		a.TypeID == fcVariantBoolType ||
		a.TypeID == fcVariantStringType {
		return a.Impl
	}

	if a.TypeID == fcVariantInt64Type {
		return Int64(a.Impl.(int64))
	}

	if a.TypeID == fcVariantUint64Type {
		return Uint64(a.Impl.(uint64))
	}

	if a.TypeID == fcVariantArrayType {
		return a.Impl.(fcVariantArray).ToNative()
	}

	if a.TypeID == fcVariantObjectType {
		return a.Impl.(fcVariantObject).ToNative()
	}

	panic(fmt.Errorf("not implemented for %s yet", fcVariantBlobType))
}

// MustAsUint64 casts the underlying `impl` as a `uint64` type, panics if not of the correct type.
func (a fcVariant) MustAsUint64() uint64 {
	return a.Impl.(uint64)
}

// MustAsString casts the underlying `impl` as a `string` type, panics if not of the correct type.
func (a fcVariant) MustAsString() string {
	return a.Impl.(string)
}

// MustAsObject casts the underlying `impl` as a `fcObject` type, panics if not of the correct type.
func (a fcVariant) MustAsObject() fcVariantObject {
	return a.Impl.(fcVariantObject)
}

func (a *fcVariant) UnmarshalBinary(decoder *Decoder) error {
	typeID, err := decoder.ReadUvarint32()
	if err != nil {
		return fmt.Errorf("unable to read fc variant type ID: %s", err)
	}

	if typeID > uint32(fcVariantBlobType) {
		return fmt.Errorf("invalid fc variant type ID, should have been lower than or equal to %d", fcVariantBlobType)
	}

	a.TypeID = fcVariantType(typeID)
	if a.TypeID == fcVariantNullType {
		// There is probably no bytes to read here, but it's not super clear
		a.Impl = nil
		return nil
	}

	if a.TypeID == fcVariantInt64Type {
		if a.Impl, err = decoder.ReadInt64(); err != nil {
			return fmt.Errorf("unable to read int64 fc variant: %s", err)
		}
	} else if a.TypeID == fcVariantUint64Type {
		if a.Impl, err = decoder.ReadUint64(); err != nil {
			return fmt.Errorf("unable to read uint64 fc variant: %s", err)
		}
	} else if a.TypeID == fcVariantDoubleType {
		if a.Impl, err = decoder.ReadFloat64(); err != nil {
			return fmt.Errorf("unable to read double fc variant: %s", err)
		}
	} else if a.TypeID == fcVariantBoolType {
		if a.Impl, err = decoder.ReadBool(); err != nil {
			return fmt.Errorf("unable to read bool fc variant: %s", err)
		}
	} else if a.TypeID == fcVariantStringType {
		if a.Impl, err = decoder.ReadString(); err != nil {
			return fmt.Errorf("unable to read string fc variant: %s", err)
		}
	} else if a.TypeID == fcVariantArrayType {
		out := fcVariantArray(nil)
		if err = decoder.Decode(&out); err != nil {
			return fmt.Errorf("unable to read fc array variant: %s", err)
		}
		a.Impl = out
	} else if a.TypeID == fcVariantObjectType {
		out := fcVariantObject{}
		if err = decoder.Decode(&out); err != nil {
			return fmt.Errorf("unable to read fc object variant: %s", err)
		}
		a.Impl = out
	} else if a.TypeID == fcVariantBlobType {
		// FIXME: This one is really not clear what the output format looks like, do we even need an object for it?
		var out fcVariantBlob
		if err = decoder.Decode(&out); err != nil {
			return fmt.Errorf("unable to read fc blob variant: %s", err)
		}
		a.Impl = out
	}

	return nil
}

type fcVariantArray []fcVariant

func (o fcVariantArray) ToNative() interface{} {
	out := make([]interface{}, len(o))
	for i, element := range o {
		out[i] = element.ToNative()
	}

	return out
}

func (o *fcVariantArray) UnmarshalBinary(decoder *Decoder) error {
	elementCount, err := decoder.ReadUvarint64()
	if err != nil {
		return fmt.Errorf("unable to read length: %s", err)
	}

	array := make([]fcVariant, elementCount)
	for i := uint64(0); i < elementCount; i++ {
		err := decoder.Decode(&array[i])
		if err != nil {
			return fmt.Errorf("unable to read elememt at index %d: %s", i, err)
		}
	}

	*o = fcVariantArray(array)
	return nil
}

type fcVariantObject map[string]fcVariant

func (o fcVariantObject) ToNative() map[string]interface{} {
	out := map[string]interface{}{}
	for key, value := range o {
		out[key] = value.ToNative()
	}

	return out
}

func (o fcVariantObject) validateFields(nameToType map[string]fcVariantType) error {
	for key, fcType := range nameToType {
		if len(key) <= 0 {
			continue
		}

		optional := false
		if string(key[0]) == "?" {
			key = key[1:]
			optional = true
		}

		actualType := o[key].TypeID
		if optional && actualType == fcVariantNullType {
			continue
		}

		if !optional && actualType == fcVariantNullType {
			return fmt.Errorf("field %q of type %s is required but actual type is null", key, fcType)
		}

		if actualType != fcType {
			return fmt.Errorf("field %q should be a variant of type %s, got %s", key, fcType, actualType)
		}
	}

	return nil
}

func (o *fcVariantObject) UnmarshalBinary(decoder *Decoder) error {
	elementCount, err := decoder.ReadUvarint64()
	if err != nil {
		return fmt.Errorf("unable to read length: %s", err)
	}

	mappings := make(map[string]fcVariant, elementCount)
	for i := uint64(0); i < elementCount; i++ {
		key, err := decoder.ReadString()
		if err != nil {
			return fmt.Errorf("unable to read key of elememt at index %d: %s", i, err)
		}

		variant := fcVariant{}
		err = decoder.Decode(&variant)
		if err != nil {
			return fmt.Errorf("unable to read value of elememt with key %s at index %d: %s", key, i, err)
		}

		mappings[key] = variant
	}

	*o = fcVariantObject(mappings)
	return nil
}

// FIXME: This one I'm unsure, is this correct at all?
type fcVariantBlob Blob

func (o *fcVariantBlob) UnmarshalBinary(decoder *Decoder) error {
	var blob Blob
	err := decoder.Decode(&blob)
	if err != nil {
		return err
	}

	*o = fcVariantBlob(blob)
	return nil
}
