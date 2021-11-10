package ojson

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Value represents a JSON value unmarshaled from a string that maintains
// key ordering in its nested component objects. Calling json.Unmarshal()
// on it behaves similarly to calling json.Unmarshal() on interface{}, except
// that objects are unmarshaled to *ojson.Object, which maintains key ordering,
// instead of map[string]interface{}, which doesn't.
type Value struct {
	V interface{}
}

var _ json.Unmarshaler = &Value{}
var _ json.Marshaler = Value{}
var _ sql.Scanner = &Value{}
var _ driver.Valuer = Value{}

// Object represents a JSON object that maintains key ordering.
type Object struct {
	keyOrder []string
	values   map[string]interface{}
}

var _ json.Marshaler = Object{}

func NewObject() *Object {
	return &Object{
		keyOrder: make([]string, 0),
		values:   make(map[string]interface{}),
	}
}

func (o *Object) Get(k string) (interface{}, bool) {
	v, ok := o.values[k]
	return v, ok
}

func (o *Object) Set(k string, v interface{}) {
	// Use original order if inserting twice.
	if _, ok := o.values[k]; !ok {
		o.keyOrder = append(o.keyOrder, k)
	}
	o.values[k] = v
}

func (o *Object) KeyOrder() []string {
	return o.keyOrder
}

// SetAndReturn is equivalent to Set, while returning a pointer to the Object.
// Primarily used for creating Objects more easily, by allowing chaining of
// multiple SetAndReturns.
func (o *Object) SetAndReturn(k string, v interface{}) *Object {
	o.Set(k, v)
	return o
}

func (o Object) MarshalJSON() ([]byte, error) {
	b := new(bytes.Buffer)
	b.WriteString("{")
	for i, k := range o.keyOrder {
		if i > 0 {
			b.WriteString(",")
		}
		if err := json.NewEncoder(b).Encode(k); err != nil {
			return nil, err
		}
		b.WriteString(":")
		v, _ := o.Get(k)
		b2, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		b.Write(b2)
	}
	b.WriteString("}")
	return b.Bytes(), nil
}

func (v Value) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.V)
}

func (v *Value) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}
	return v.UnmarshalJSON(source)
}

func (v *Value) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	oj, d, err := unmarshal(dec)
	if d != 0 {
		return errors.New("unexpected delimiter")
	}
	v.V = oj
	return err
}

func NewValueFromJSON(s string) (Value, error) {
	var v Value
	if err := v.UnmarshalJSON([]byte(s)); err != nil {
		return v, err
	}
	return v, nil
}

func MustNewValueFromJSON(s string) Value {
	if v, err := NewValueFromJSON(s); err != nil {
		panic(err)
	} else {
		return v
	}
}

// unmarshal consumes from dec to decode the next chunk of JSON. It either
// returns a JSON value corresponding to the result of a successful parse, or
// a delimiter token if that is the next value in the decoder (needed to
// correctly parse the ending ']' character of a JSON array)
func unmarshal(dec *json.Decoder) (interface{}, json.Delim, error) {
	var o interface{}
	t, err := dec.Token()
	if err != nil {
		return nil, 0, err
	}
	switch v := t.(type) {
	case json.Delim:
		switch v {
		case '{':
			obj, err := unmarshalObject(dec)
			if err != nil {
				return nil, 0, err
			}
			o = obj

		case '[':
			arr, err := unmarshalArray(dec)
			if err != nil {
				return nil, 0, err
			}
			o = arr

		default:
			return nil, v, nil
		}

	case float64, string, bool, nil:
		o = v

	default:
		return nil, 0, errors.New("unexpected type")
	}
	return o, 0, nil
}

func unmarshalArray(dec *json.Decoder) ([]interface{}, error) {
	arr := make([]interface{}, 0)
	for {
		o, d, err := unmarshal(dec)
		if err != nil {
			return arr, err
		}
		switch d {
		case ']':
			return arr, nil
		case 0:
			arr = append(arr, o)
		default:
			return arr, errors.New("unexpected delimiter (expecting ])")
		}
	}
}

func unmarshalObject(dec *json.Decoder) (*Object, error) {
	obj := NewObject()
	for {
		t, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch v := t.(type) {
		case json.Delim:
			if v == '}' {
				return obj, nil
			} else {
				return nil, errors.New("unexpected delimiter (expecting })")
			}

		case string:
			o, d, err := unmarshal(dec)
			if err != nil {
				return nil, err
			}
			if d != 0 {
				return nil, errors.New("unexpected delimiter")
			}
			obj.Set(v, o)

		default:
			return nil, errors.New("unexpected token")
		}
	}
}
