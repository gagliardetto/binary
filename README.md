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
