package jsonstruct

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoType(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		values                []interface{}
		expectedObservedValue *ObservedValue
		generatorOptions      []GeneratorOption
		expectedGoType        string
		expectedImports       map[string]bool
	}{
		{
			name: "slice_empty",
			values: []interface{}{
				[]interface{}{},
			},
			expectedObservedValue: &ObservedValue{
				Observations:     1,
				Array:            1,
				AllArrayElements: &ObservedValue{},
			},
			expectedGoType: "[]interface{}",
		},
		{
			name: "slice_bool",
			values: []interface{}{
				[]interface{}{
					false,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Array:        1,
				AllArrayElements: &ObservedValue{
					Observations: 1,
					Bool:         1,
				},
			},
			expectedGoType: "[]bool",
		},
		{
			name: "bool",
			values: []interface{}{
				false,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Bool:         1,
			},
			expectedGoType: "bool",
		},
		{
			name: "bool_and_null",
			values: []interface{}{
				false,
				nil,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				Bool:         1,
				Null:         1,
			},
			expectedGoType: "*bool",
		},
		{
			name: "float64",
			values: []interface{}{
				0.0,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Float64:      1,
			},
			expectedGoType: "float64",
		},
		{
			name: "float64_and_null",
			values: []interface{}{
				0.0,
				nil,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				Float64:      1,
				Null:         1,
			},
			expectedGoType: "*float64",
		},
		{
			name: "int",
			values: []interface{}{
				0,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Int:          1,
			},
			expectedGoType: "int",
		},
		{
			name: "int_and_null",
			values: []interface{}{
				0,
				nil,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				Int:          1,
				Null:         1,
			},
			expectedGoType: "*int",
		},
		{
			name: "float64_and_int",
			values: []interface{}{
				0.0,
				0,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				Float64:      1,
				Int:          1,
			},
			expectedGoType: "float64",
		},
		{
			name: "float64_and_int_and_null",
			values: []interface{}{
				0.0,
				0,
				nil,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 3,
				Float64:      1,
				Int:          1,
				Null:         1,
			},
			expectedGoType: "*float64",
		},
		{
			name: "object_empty",
			values: []interface{}{
				map[string]interface{}{},
			},
			expectedObservedValue: &ObservedValue{
				Observations:         1,
				Object:               1,
				ObjectPropertyValues: map[string]*ObservedValue{},
			},
			expectedGoType: "struct{}",
		},
		{
			name: "object_and_nil",
			values: []interface{}{
				map[string]interface{}{},
				nil,
			},
			expectedObservedValue: &ObservedValue{
				Observations:         2,
				Null:                 1,
				Object:               1,
				ObjectPropertyValues: map[string]*ObservedValue{},
			},
			expectedGoType: "struct{}",
		},
		{
			name: "object_simple",
			values: []interface{}{
				map[string]interface{}{
					"key": false,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Object:       1,
				ObjectPropertyValues: map[string]*ObservedValue{
					"key": {
						Observations: 1,
						Bool:         1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 1,
					Bool:         1,
				},
			},
			expectedGoType: "struct {\nKey bool `json:\"key\"`\n}",
		},
		{
			name: "object_unparseable_properties_skip",
			values: []interface{}{
				map[string]interface{}{
					"key with spaces": false,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Object:       1,
				ObjectPropertyValues: map[string]*ObservedValue{
					"key with spaces": {
						Observations: 1,
						Bool:         1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 1,
					Bool:         1,
				},
			},
			expectedGoType: "struct {\n// \"key with spaces\" cannot be unmarshalled into a struct field by encoding/json.\n}",
		},
		{
			name: "object_unparseable_properties",
			values: []interface{}{
				map[string]interface{}{
					"key with spaces":         false,
					"another key with spaces": true,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Object:       1,
				ObjectPropertyValues: map[string]*ObservedValue{
					"key with spaces": {
						Observations: 1,
						Bool:         1,
					},
					"another key with spaces": {
						Observations: 1,
						Bool:         1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 2,
					Bool:         2,
				},
			},
			generatorOptions: []GeneratorOption{
				WithSkipUnparseableProperties(false),
			},
			expectedGoType: "map[string]bool",
		},
		{
			name: "object_unparseable_properties_variable_values",
			values: []interface{}{
				map[string]interface{}{
					"key with spaces":         false,
					"another key with spaces": 0,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Object:       1,
				ObjectPropertyValues: map[string]*ObservedValue{
					"key with spaces": {
						Observations: 1,
						Bool:         1,
					},
					"another key with spaces": {
						Observations: 1,
						Int:          1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 2,
					Bool:         1,
					Int:          1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithSkipUnparseableProperties(false),
			},
			expectedGoType: "map[string]interface{}",
		},
		{
			name: "object_kebab_case",
			values: []interface{}{
				map[string]interface{}{
					"kebab-case": true,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Object:       1,
				ObjectPropertyValues: map[string]*ObservedValue{
					"kebab-case": {
						Observations: 1,
						Bool:         1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 1,
					Bool:         1,
				},
			},
			expectedGoType: "struct {\nKebabCase bool `json:\"kebab-case\"`\n}",
		},
		{
			name: "string",
			values: []interface{}{
				"",
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				String:       1,
			},
			expectedGoType: "string",
		},
		{
			name: "string_and_null",
			values: []interface{}{
				"",
				nil,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				String:       1,
				Null:         1,
			},
			expectedGoType: "*string",
		},
		{
			name: "time",
			values: []interface{}{
				"1985-04-12T23:20:50.52Z",
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				String:       1,
				Time:         1,
			},
			expectedGoType: "time.Time",
			expectedImports: map[string]bool{
				"time": true,
			},
		},
		{
			name: "time_and_null",
			values: []interface{}{
				"1985-04-12T23:20:50.52Z",
				nil,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				Null:         1,
				String:       1,
				Time:         1,
			},
			expectedGoType: "*time.Time",
			expectedImports: map[string]bool{
				"time": true,
			},
		},
		{
			name: "time_and_string",
			values: []interface{}{
				"1985-04-12T23:20:50.52Z",
				"",
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				String:       2,
				Time:         1,
			},
			expectedGoType: "string",
		},
		{
			name: "time_and_string_and_null",
			values: []interface{}{
				"1985-04-12T23:20:50.52Z",
				"",
				nil,
			},
			expectedObservedValue: &ObservedValue{
				Observations: 3,
				Null:         1,
				String:       2,
				Time:         1,
			},
			expectedGoType: "*string",
		},
		{
			name: "custom_fieldnamer",
			values: []interface{}{
				map[string]interface{}{
					"gpsAltitude": 0,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Object:       1,
				ObjectPropertyValues: map[string]*ObservedValue{
					"gpsAltitude": {
						Observations: 1,
						Int:          1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 1,
					Int:          1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithFieldNamer(&AbbreviationHandlingFieldNamer{
					Abbreviations: map[string]bool{
						"GPS": true,
					},
				}),
			},
			expectedGoType: "struct {\nGPSAltitude int `json:\"gpsAltitude\"`\n}",
		},
		{
			name: "omitempty_always",
			values: []interface{}{
				map[string]interface{}{
					"key": 0,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 1,
				Object:       1,
				ObjectPropertyValues: map[string]*ObservedValue{
					"key": {
						Observations: 1,
						Int:          1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 1,
					Int:          1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithOmitEmpty(OmitEmptyAlways),
			},
			expectedGoType: "struct {\nKey int `json:\"key,omitempty\"`\n}",
		},
		{
			name: "omitempty_never",
			values: []interface{}{
				map[string]interface{}{
					"key": 0,
				},
				map[string]interface{}{},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				Object:       2,
				ObjectPropertyValues: map[string]*ObservedValue{
					"key": {
						Observations: 1,
						Int:          1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 1,
					Int:          1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithOmitEmpty(OmitEmptyNever),
			},
			expectedGoType: "struct {\nKey int `json:\"key\"`\n}",
		},
		{
			name: "omitempty_auto",
			values: []interface{}{
				map[string]interface{}{
					"key1": 0,
					"key2": 0,
				},
				map[string]interface{}{
					"key1": 0,
				},
			},
			expectedObservedValue: &ObservedValue{
				Observations: 2,
				Object:       2,
				ObjectPropertyValues: map[string]*ObservedValue{
					"key1": {
						Observations: 2,
						Int:          2,
					},
					"key2": {
						Observations: 1,
						Int:          1,
					},
				},
				AllObjectPropertyValues: &ObservedValue{
					Observations: 3,
					Int:          3,
				},
			},
			generatorOptions: []GeneratorOption{
				WithOmitEmpty(OmitEmptyAuto),
			},
			expectedGoType: "struct {\nKey1 int `json:\"key1\"`\nKey2 int `json:\"key2,omitempty\"`\n}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualObservedValue := &ObservedValue{}
			for _, value := range tc.values {
				actualObservedValue = actualObservedValue.Merge(value)
			}
			assert.Equal(t, tc.expectedObservedValue, actualObservedValue)
			actualImports := make(map[string]bool)
			goType, _ := NewGenerator(tc.generatorOptions...).GoType(actualObservedValue, 0, actualImports)
			assert.Equal(t, tc.expectedGoType, goType)
			if len(tc.expectedImports) == 0 {
				assert.Empty(t, actualImports)
			} else {
				assert.Equal(t, tc.expectedImports, actualImports)
			}
		})
	}
}

func TestObserveGoCode(t *testing.T) {
	for _, tc := range []struct {
		name              string
		json              string
		wantErr           bool
		generatorOptions  []GeneratorOption
		expectedGoCodeStr string
	}{
		{
			name:    "error",
			json:    `"`,
			wantErr: true,
		},
		{
			name: "empty",
			json: ``,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T interface{}\n",
		},
		{
			name: "bool",
			json: `true`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T bool\n",
		},
		{
			name: "int",
			json: `0`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T int\n",
		},
		{
			name: "float64",
			json: `0.0`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T float64\n",
		},
		{
			name: "comments_and_names",
			json: `true`,
			generatorOptions: []GeneratorOption{
				WithPackageComment("package demo."),
				WithPackageName("demo"),
				WithTypeComment("MyType is my type."),
				WithTypeName("MyType"),
			},
			expectedGoCodeStr: "" +
				"// package demo.\n" +
				"package demo\n" +
				"\n" +
				"// MyType is my type.\n" +
				"type MyType bool\n",
		},
		{
			name: "time",
			json: `"1985-04-12T23:20:50.52Z"`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"import (\n" +
				"\t\"time\"\n" +
				")\n" +
				"\n" +
				"type T time.Time\n",
		},
		{
			name: "auto_omitempty",
			json: "" +
				`{"intKey":0,"boolKey":true}` +
				`{"intKey":0}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tBoolKey bool `json:\"boolKey,omitempty\"`\n" +
				"\tIntKey  int  `json:\"intKey\"`\n" +
				"}\n",
		},
		{
			name: "slice_in_object",
			json: "" +
				`{"slice":[]}` +
				`{"slice":[0]}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tSlice []int `json:\"slice\"`\n" +
				"}\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			observedValue, err := Observe(bytes.NewBufferString(tc.json))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			goCode, err := NewGenerator(tc.generatorOptions...).GoCode(observedValue)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedGoCodeStr, string(goCode))
		})
	}
}
