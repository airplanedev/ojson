# ojson
Go library for ordered JSON values.

## Motivation
Go's JSON (de-)serialization doesn't maintain the order of keys. Despite this being the expected behavior according to the JSON spec, some legacy JSON implementations expect key order to be preserved, especially since JSON is a commonly-used serialization protocol for data between different systems. 

Thus, this library allows handling of ordered JSON data. Calling `json.Unmarshal()` on an `ojson.Value` behaves similarly to calling `json.Unmarshal()` on an `interface{}`, except that objects are unmarshaled to `*ojson.Object`, which is a struct containing a `map[string]interface{}` plus a key ordering. Furthermore, calling `json.Marshal()` on an `ojson.Object` outputs JSON that is serialized with the declared key ordering.

## Example
```go
package main

import (
  "encoding/json"

  "github.com/airplanedev/ojson"
)

func main() {
  var v ojson.Value
  b := []byte(`{"b":[],"c":true,"a":{"2":1,"1":{}}}`)
  json.Unmarshal(b, &v)

  // Here, s.([]byte) should be equal to b, preserving the order of keys in the relevant JSON objects.
  s, _ := json.Marshal(v)
}
```
