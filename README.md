# binary


### Borsh

#### Decoding borsh

```golang
 dec := bin.NewBorshDecoder(data)
 var meta token_metadata.Metadata
 err = dec.Decode(&meta)
 if err != nil {
   panic(err)
 }
```

#### Encoding borsh

```golang
buf := new(bytes.Buffer)
enc := bin.NewBorshEncoder(buf)
err := enc.Encode(meta)
if err != nil {
  panic(err)
}
// fmt.Print(buf.Bytes())
```

### Optional Types

```
type Person struct {
	Name string
	Age  uint8 `bin:"optional"`
}
```

Rust equivalent:
```
struct Person {
    name: String,
    age: Option<u8>
}
```

### Enum Types

```
type MyEnum struct {
	Enum  bin.BorshEnum `borsh_enum:"true"`
	One   bin.EmptyVariant
	Two   uint32
	Three int16
}
```

Rust equivalent:
```
enum MyEnum {
    One,
    Two(u32),
    Three(i16),
}
```

