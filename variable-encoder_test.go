// Copyright (c) 2024 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package typst_test

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Dadido3/go-typst"
	"github.com/google/go-cmp/cmp"
)

func TestMarshalVariable(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    []byte
		wantErr bool
	}{
		{"nil", nil, []byte(`none`), false},
		{"string", "Hey\nThere!", []byte(`"Hey\nThere!"`), false},
		{"int", -123, []byte(`{-123}`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := typst.MarshalVariable(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalVariable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}

type VariableMarshalerType []byte

func (v VariableMarshalerType) MarshalTypstVariable() ([]byte, error) {
	result := append([]byte{'"'}, v...)
	result = append(result, '"')

	return result, nil
}

type VariableMarshalerTypePointer []byte

var variableMarshalerTypePointer = VariableMarshalerTypePointer("test")
var variableMarshalerTypePointerNil = VariableMarshalerTypePointer(nil)

func (v *VariableMarshalerTypePointer) MarshalTypstVariable() ([]byte, error) {
	if v != nil {
		result := append([]byte{'"'}, *v...)
		result = append(result, '"')

		return result, nil
	}

	return nil, fmt.Errorf("no data")
}

type TextMarshalerType []byte

func (v TextMarshalerType) MarshalText() ([]byte, error) {
	return v, nil
}

type TextMarshalerTypePointer []byte

var textMarshalerTypePointer = TextMarshalerTypePointer("test")
var textMarshalerTypePointerNil = TextMarshalerTypePointer(nil)

func (v *TextMarshalerTypePointer) MarshalText() ([]byte, error) {
	if v != nil {
		return *v, nil
	}

	return nil, fmt.Errorf("no data")
}

func TestVariableEncoder(t *testing.T) {

	tests := []struct {
		name    string
		params  any
		wantErr bool
		want    string
	}{
		{"nil", nil, false, "none"},
		{"bool false", false, false, "false"},
		{"bool true", true, false, "true"},
		{"int", int(123), false, "123"},
		{"int8", int8(123), false, "123"},
		{"int16", int16(123), false, "123"},
		{"int32", int32(123), false, "123"},
		{"int64", int64(123), false, "123"},
		{"int negative", int(-123), false, "{-123}"},
		{"int8 negative", int8(-123), false, "{-123}"},
		{"int16 negative", int16(-123), false, "{-123}"},
		{"int32 negative", int32(-123), false, "{-123}"},
		{"int64 negative", int64(-123), false, "{-123}"},
		{"uint", uint(123), false, "123"},
		{"uint8", uint8(123), false, "123"},
		{"uint16", uint16(123), false, "123"},
		{"uint32", uint32(123), false, "123"},
		{"uint64", uint64(123), false, "123"},
		{"float32", float32(1), false, "1e+00"},
		{"float64", float64(1), false, "1e+00"},
		{"float32 negative", float32(-1), false, "{-1e+00}"},
		{"float64 negative", float64(-1), false, "{-1e+00}"},
		{"float64 nan", float64(math.NaN()), false, "float.nan"},
		{"float64 +inf", float64(math.Inf(1)), false, "float.inf"},
		{"float64 -inf", float64(math.Inf(-1)), false, "{-float.inf}"},
		{"string", "Hey!", false, `"Hey!"`},
		{"string escaped", "Hey!😀 \"This is quoted\"\nNew line!\tAnd a tab", false, `"Hey!😀 \"This is quoted\"\nNew line!\tAnd a tab"`},
		{"struct", struct {
			Foo string
			Bar int
		}{"Hey!", 12345}, false, "(\n  Foo: \"Hey!\",\n  Bar: 12345,\n)"},
		{"struct empty", struct{}{}, false, "()"},
		{"struct empty pointer", (*struct{})(nil), false, "none"},
		{"map string string", map[string]string{"Foo": "Bar", "Foo2": "Electric Foogaloo"}, false, "(\n  Foo: \"Bar\",\n  Foo2: \"Electric Foogaloo\",\n)"},
		{"map string string empty", map[string]string{}, false, "()"},
		{"map string string nil", map[string]string(nil), false, "()"},
		{"string array", [5]string{"Foo", "Bar"}, false, `("Foo", "Bar", "", "", "")`},
		{"string array 1", [1]string{"Foo"}, false, `("Foo",)`},
		{"string slice", []string{"Foo", "Bar"}, false, `("Foo", "Bar")`},
		{"string slice 1", []string{"Foo"}, false, `("Foo",)`},
		{"string slice empty", []string{}, false, `()`},
		{"string slice nil", []string(nil), false, `()`},
		{"string slice pointer", &[]string{"Foo", "Bar"}, false, `("Foo", "Bar")`},
		{"int slice", []int{1, 2, 3, 4, 5}, false, `(1, 2, 3, 4, 5)`},
		{"int slice negative", []int{1, -2, 3, -4, 5}, false, `(1, {-2}, 3, {-4}, 5)`},
		{"byte slice", []byte{1, 2, 3, 4, 5}, false, `bytes((1, 2, 3, 4, 5))`},
		{"byte slice 1", []byte{1}, false, `bytes((1,))`},
		{"MarshalTypstVariable value", VariableMarshalerType("test"), false, `"test"`},
		{"MarshalTypstVariable value nil", VariableMarshalerType(nil), false, `""`},
		{"MarshalTypstVariable pointer", &variableMarshalerTypePointer, false, `"test"`},
		{"MarshalTypstVariable pointer nil", &variableMarshalerTypePointerNil, false, `""`},
		{"MarshalTypstVariable nil pointer", struct{ A *VariableMarshalerTypePointer }{nil}, true, ``},
		{"MarshalText value", TextMarshalerType("test"), false, `"test"`},
		{"MarshalText value nil", TextMarshalerType(nil), false, `""`},
		{"MarshalText pointer", &textMarshalerTypePointer, false, `"test"`},
		{"MarshalText pointer nil", &textMarshalerTypePointerNil, false, `""`},
		{"MarshalText nil pointer", struct{ A *TextMarshalerTypePointer }{nil}, true, ``},
		{"time.Time", time.Date(2024, 12, 14, 12, 34, 56, 0, time.UTC), false, `datetime(year: 2024, month: 12, day: 14, hour: 12, minute: 34, second: 56)`},
		{"time.Time pointer", &[]time.Time{time.Date(2024, 12, 14, 12, 34, 56, 0, time.UTC)}[0], false, `datetime(year: 2024, month: 12, day: 14, hour: 12, minute: 34, second: 56)`},
		{"time.Time pointer nil", (*time.Time)(nil), false, `none`},
		{"time.Duration", 60 * time.Second, false, `duration(seconds: 60)`},
		{"time.Duration pointer", &[]time.Duration{60 * time.Second}[0], false, `duration(seconds: 60)`},
		{"time.Duration pointer nil", (*time.Duration)(nil), false, `none`},
		{"time.Duration negative", -60 * time.Second, false, `duration(seconds: -60)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result bytes.Buffer
			vEnc := typst.NewVariableEncoder(&result)

			err := vEnc.Encode(tt.params)
			switch {
			case err != nil && !tt.wantErr:
				t.Fatalf("Failed to encode typst variables: %v", err)
			case err == nil && tt.wantErr:
				t.Fatalf("Expected error, but got none")
			}

			if !tt.wantErr && !cmp.Equal(result.String(), tt.want) {
				t.Errorf("Got the following diff in output: %s", cmp.Diff(tt.want, result.String()))
			}

			// Compile to test parsing.
			if !tt.wantErr {
				typstCLI := typst.CLI{}
				input := strings.NewReader("#" + result.String())
				var output bytes.Buffer
				if err := typstCLI.Render(input, &output, nil); err != nil {
					t.Errorf("Compilation failed: %v", err)
				}
			}
		})
	}
}
