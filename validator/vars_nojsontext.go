//go:build !go1.27 || !goexperiment.jsonv2

package validator

import (
	"encoding/json"
	"reflect"
)

// jsonTextValueType is nil when encoding/json/jsontext is not available.
var jsonTextValueType reflect.Type

func jsonTextNumber(any) (json.Number, bool) {
	return "", false
}
