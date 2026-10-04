//go:build go1.27 && goexperiment.jsonv2

package validator

import (
	"encoding/json"
	"encoding/json/jsontext"
	"reflect"
)

var jsonTextValueType = reflect.TypeFor[jsontext.Value]()

// jsonTextNumber returns v as a json.Number if v is a jsontext.Value that
// holds a JSON number. encoding/json/v2 has no UseNumber, so decoders built
// on it keep numbers as jsontext.Value to keep their precision.
func jsonTextNumber(v any) (json.Number, bool) {
	raw, ok := v.(jsontext.Value)
	if !ok || raw.Kind() != '0' {
		return "", false
	}
	return json.Number(raw), true
}
