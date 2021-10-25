package ojson

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshal(tt *testing.T) {
	for _, test := range []struct {
		serialized   string
		unserialized Value
	}{
		{
			`{"gh\"jkl":[true,123,["asdf"]],"asdf":null}`,
			Value{
				V: &Object{
					keyOrder: []string{"gh\"jkl", "asdf"},
					values: map[string]interface{}{
						"asdf": nil,
						"gh\"jkl": []interface{}{
							true,
							123.0,
							[]interface{}{
								"asdf",
							},
						},
					},
				},
			},
		},
		{
			`{"b":{"c":1,"d":2},"a":{"d":2,"c":1}}`,
			Value{
				V: &Object{
					keyOrder: []string{"b", "a"},
					values: map[string]interface{}{
						"a": &Object{
							keyOrder: []string{"d", "c"},
							values: map[string]interface{}{
								"c": 1.0,
								"d": 2.0,
							},
						},
						"b": &Object{
							keyOrder: []string{"c", "d"},
							values: map[string]interface{}{
								"c": 1.0,
								"d": 2.0,
							},
						},
					},
				},
			},
		},
	} {
		tt.Run(test.serialized, func(t *testing.T) {
			require := require.New(t)
			s, err := json.Marshal(test.unserialized)
			require.NoError(err)
			require.Equal([]byte(test.serialized), s)

			var o Value
			require.NoError(json.Unmarshal([]byte(test.serialized), &o))
			require.Equal(test.unserialized, o)
		})
		tt.Run("scan and value "+test.serialized, func(t *testing.T) {
			require := require.New(t)
			s, err := test.unserialized.Value()
			require.NoError(err)
			require.Equal([]byte(test.serialized), s.([]byte))

			var o Value
			require.NoError(o.Scan([]byte(test.serialized)))
			require.Equal(test.unserialized, o)
		})
	}
}

// TestMarshalValidJson tests that marshaling an ojson object with all
// keys already sorted gets the equivalent result to marshaling a similar
// non-ojson object.
func TestMarshalValidJson(tt *testing.T) {
	for _, test := range []struct {
		name string
		j    interface{}
		oj   Value
	}{
		{
			name: "nested sorted objects",
			j: map[string]interface{}{
				"a": map[string]interface{}{
					"c": 1.0,
					"d": 2.0,
				},
				"b": map[string]interface{}{
					"f": 1.0,
					"e": 2.0,
				},
			},
			oj: Value{
				V: &Object{
					keyOrder: []string{"a", "b"},
					values: map[string]interface{}{
						"a": &Object{
							keyOrder: []string{"c", "d"},
							values: map[string]interface{}{
								"c": 1.0,
								"d": 2.0,
							},
						},
						"b": &Object{
							keyOrder: []string{"e", "f"},
							values: map[string]interface{}{
								"f": 1.0,
								"e": 2.0,
							},
						},
					},
				},
			},
		},
	} {
		tt.Run(test.name, func(t *testing.T) {
			require := require.New(t)
			s, err := json.Marshal(test.oj)
			require.NoError(err)
			s2, err := json.Marshal(test.j)
			require.NoError(err)
			require.Equal(s, s2)
		})
	}
}
