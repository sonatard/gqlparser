//go:build go1.27 && goexperiment.jsonv2

package validator_test

import (
	"encoding/json"
	"encoding/json/jsontext"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"
)

// TestValidateVarsJSONTextNumbers checks that a JSON number in a
// jsontext.Value is validated as the equivalent json.Number, and that the
// jsontext.Value is kept wherever the json.Number would be kept.
func TestValidateVarsJSONTextNumbers(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "vars.graphql",
		Input: mustReadFile("./testdata/vars.graphql"),
	})

	// sliceOf returns a slice of v's type that holds v, as the validator
	// coerces a single value to a list.
	sliceOf := func(v any) any {
		s := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(v)), 0, 1)
		return reflect.Append(s, reflect.ValueOf(v)).Interface()
	}

	cases := []struct {
		name    string
		query   string
		value   func(num func(string) any) any
		want    func(num func(string) any) any
		wantErr string
	}{
		{
			name:  "Int",
			query: `query foo($var: Int) { optionalIntArg(i: $var) }`,
			value: func(num func(string) any) any { return num("10") },
			want:  func(func(string) any) any { return int64(10) },
		},
		{
			name:    "Int with a fraction",
			query:   `query foo($var: Int) { optionalIntArg(i: $var) }`,
			value:   func(num func(string) any) any { return num("1.5") },
			wantErr: "input: variable.var cannot use value 0 as Int",
		},
		{
			name:  "Float",
			query: `query foo($var: Float!) { floatArg(i: $var) }`,
			value: func(num func(string) any) any { return num("10.5") },
			want:  func(func(string) any) any { return 10.5 },
		},
		{
			name:  "ID",
			query: `query foo($var: ID!) { idArg(i: $var) }`,
			value: func(num func(string) any) any { return num("5") },
			want:  func(num func(string) any) any { return num("5") },
		},
		{
			name:  "String",
			query: `query foo($var: String) { stringArg(i: $var) }`,
			value: func(num func(string) any) any { return num("5") },
			want:  func(num func(string) any) any { return num("5") },
		},
		{
			name:  "custom scalar",
			query: `query foo($var: Custom!) { scalarArg(i: $var) }`,
			value: func(num func(string) any) any { return num("5") },
			want:  func(num func(string) any) any { return num("5") },
		},
		{
			name:    "Boolean",
			query:   `query foo($var: Boolean!) { boolArg(i: $var) }`,
			value:   func(num func(string) any) any { return num("1") },
			wantErr: "input: variable.var cannot use string as Boolean",
		},
		{
			name:  "coerced to a list",
			query: `query foo($var: [Int!]) { intArrayArg(i: $var) }`,
			value: func(num func(string) any) any { return num("5") },
			want:  func(num func(string) any) any { return sliceOf(num("5")) },
		},
		{
			name:  "list",
			query: `query foo($var: [Int]) { intArrayArg(i: $var) }`,
			value: func(num func(string) any) any { return []any{num("1"), num("2")} },
			want:  func(num func(string) any) any { return []any{num("1"), num("2")} },
		},
		{
			name:    "list with a fraction",
			query:   `query foo($var: [Int]) { intArrayArg(i: $var) }`,
			value:   func(num func(string) any) any { return []any{num("1.5")} },
			wantErr: "input: variable.var[0] cannot use string as Int",
		},
		{
			name:  "input object field coerced to a list",
			query: `query foo($var: [CustomType]) { typeArrayArg(i: $var) }`,
			value: func(num func(string) any) any {
				return []any{map[string]any{"and": num("5")}}
			},
			want: func(num func(string) any) any {
				return []any{map[string]any{"and": sliceOf(num("5"))}}
			},
		},
	}

	numbers := map[string]func(string) any{
		"json.Number":    func(s string) any { return json.Number(s) },
		"jsontext.Value": func(s string) any { return jsontext.Value(s) },
	}

	for _, tc := range cases {
		for numName, num := range numbers {
			t.Run(tc.name+"/"+numName, func(t *testing.T) {
				//nolint:staticcheck
				q := gqlparser.MustLoadQuery(schema, tc.query)
				vars, gerr := validator.VariableValues(
					schema,
					q.Operations.ForName(""),
					map[string]any{"var": tc.value(num)},
				)
				if tc.wantErr != "" {
					require.EqualError(t, gerr, tc.wantErr)
					return
				}
				require.NoError(t, gerr)
				require.Equal(t, tc.want(num), vars["var"])
			})
		}
	}
}
