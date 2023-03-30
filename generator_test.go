package jsonstruct

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestGoType(t *testing.T) {
	for _, tc := range []struct {
		name             string
		values           []any
		expectedValue    *value
		generatorOptions []GeneratorOption
		expectedGoType   string
		expectedImports  map[string]struct{}
	}{
		{
			name: "slice_empty",
			values: []any{
				[]any{},
			},
			expectedValue: &value{
				observations:  1,
				emptys:        1,
				arrays:        1,
				arrayElements: &value{},
			},
			expectedGoType: "[]any",
		},
		{
			name: "slice_bool",
			values: []any{
				[]any{
					false,
				},
			},
			expectedValue: &value{
				observations: 1,
				arrays:       1,
				arrayElements: &value{
					observations: 1,
					emptys:       1,
					bools:        1,
				},
			},
			expectedGoType: "[]bool",
		},
		{
			name: "bool_false",
			values: []any{
				false,
			},
			expectedValue: &value{
				observations: 1,
				emptys:       1,
				bools:        1,
			},
			expectedGoType: "bool",
		},
		{
			name: "bool_true",
			values: []any{
				true,
			},
			expectedValue: &value{
				observations: 1,
				bools:        1,
			},
			expectedGoType: "bool",
		},
		{
			name: "bool_and_null",
			values: []any{
				false,
				nil,
			},
			expectedValue: &value{
				observations: 2,
				emptys:       1,
				bools:        1,
				nulls:        1,
			},
			expectedGoType: "*bool",
		},
		{
			name: "float64_zero",
			values: []any{
				0.0,
			},
			expectedValue: &value{
				observations: 1,
				emptys:       1,
				float64s:     1,
			},
			expectedGoType: "float64",
		},
		{
			name: "float64_nonzero",
			values: []any{
				1.0,
			},
			expectedValue: &value{
				observations: 1,
				float64s:     1,
			},
			expectedGoType: "float64",
		},
		{
			name: "float64_and_null",
			values: []any{
				0.0,
				nil,
			},
			expectedValue: &value{
				observations: 2,
				emptys:       1,
				float64s:     1,
				nulls:        1,
			},
			expectedGoType: "*float64",
		},
		{
			name: "int_zero",
			values: []any{
				0,
			},
			expectedValue: &value{
				observations: 1,
				emptys:       1,
				ints:         1,
			},
			expectedGoType: "int",
		},
		{
			name: "int_nonzero",
			values: []any{
				1,
			},
			expectedValue: &value{
				observations: 1,
				ints:         1,
			},
			expectedGoType: "int",
		},
		{
			name: "int_and_null",
			values: []any{
				0,
				nil,
			},
			expectedValue: &value{
				observations: 2,
				emptys:       1,
				ints:         1,
				nulls:        1,
			},
			expectedGoType: "*int",
		},
		{
			name: "int32_zero",
			values: []any{
				0,
			},
			expectedValue: &value{
				observations: 1,
				emptys:       1,
				ints:         1,
			},
			generatorOptions: []GeneratorOption{
				WithIntType("int32"),
			},
			expectedGoType: "int32",
		},
		{
			name: "int64_zero",
			values: []any{
				0,
			},
			expectedValue: &value{
				observations: 1,
				emptys:       1,
				ints:         1,
			},
			generatorOptions: []GeneratorOption{
				WithIntType("int64"),
			},
			expectedGoType: "int64",
		},
		{
			name: "float64_and_int",
			values: []any{
				0.0,
				0,
			},
			expectedValue: &value{
				observations: 2,
				emptys:       2,
				float64s:     1,
				ints:         1,
			},
			expectedGoType: "float64",
		},
		{
			name: "float64_and_int_json_number",
			values: []any{
				0.0,
				0,
			},
			expectedValue: &value{
				observations: 2,
				emptys:       2,
				float64s:     1,
				ints:         1,
			},
			generatorOptions: []GeneratorOption{
				WithUseJSONNumber(true),
			},
			expectedGoType: "json.Number",
			expectedImports: map[string]struct{}{
				"encoding/json": {},
			},
		},
		{
			name: "float64_and_int_and_null",
			values: []any{
				0.0,
				0,
				nil,
			},
			expectedValue: &value{
				observations: 3,
				emptys:       2,
				float64s:     1,
				ints:         1,
				nulls:        1,
			},
			expectedGoType: "*float64",
		},
		{
			name: "float64_and_int_and_null_json_number",
			values: []any{
				0.0,
				0,
				nil,
			},
			expectedValue: &value{
				observations: 3,
				emptys:       2,
				float64s:     1,
				ints:         1,
				nulls:        1,
			},
			generatorOptions: []GeneratorOption{
				WithUseJSONNumber(true),
			},
			expectedGoType: "*json.Number",
			expectedImports: map[string]struct{}{
				"encoding/json": {},
			},
		},
		{
			name: "object_empty",
			values: []any{
				map[string]any{},
			},
			expectedValue: &value{
				observations:     1,
				emptys:           1,
				objects:          1,
				objectProperties: map[string]*value{},
			},
			expectedGoType: "struct{}",
		},
		{
			name: "object_and_null",
			values: []any{
				map[string]any{},
				nil,
			},
			expectedValue: &value{
				observations:     2,
				emptys:           1,
				nulls:            1,
				objects:          1,
				objectProperties: map[string]*value{},
			},
			expectedGoType: "*struct{}",
		},
		{
			name: "object_simple",
			values: []any{
				map[string]any{
					"key": false,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"key": {
						observations: 1,
						emptys:       1,
						bools:        1,
					},
				},
				allObjectProperties: &value{
					observations: 1,
					emptys:       1,
					bools:        1,
				},
			},
			expectedGoType: "struct {\nKey bool `json:\"key\"`\n}",
		},
		{
			name: "object_with_nested_int",
			values: []any{
				map[string]any{
					"key": 1,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"key": {
						observations: 1,
						ints:         1,
					},
				},
				allObjectProperties: &value{
					observations: 1,
					ints:         1,
				},
			},
			expectedGoType: "struct {\nKey int `json:\"key\"`\n}",
		},
		{
			name: "object_with_nested_int_and_int64_option",
			values: []any{
				map[string]any{
					"key": 1,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"key": {
						observations: 1,
						ints:         1,
					},
				},
				allObjectProperties: &value{
					observations: 1,
					ints:         1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithIntType("int64"),
			},
			expectedGoType: "struct {\nKey int64 `json:\"key\"`\n}",
		},
		{
			name: "object_unparseable_properties_skip",
			values: []any{
				map[string]any{
					"key with spaces": false,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"key with spaces": {
						observations: 1,
						emptys:       1,
						bools:        1,
					},
				},
				allObjectProperties: &value{
					observations: 1,
					emptys:       1,
					bools:        1,
				},
			},
			expectedGoType: "struct {\n// \"key with spaces\" cannot be unmarshalled into a struct field by encoding/json.\n}",
		},
		{
			name: "object_unparseable_properties",
			values: []any{
				map[string]any{
					"key with spaces":         false,
					"another key with spaces": true,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"key with spaces": {
						observations: 1,
						emptys:       1,
						bools:        1,
					},
					"another key with spaces": {
						observations: 1,
						bools:        1,
					},
				},
				allObjectProperties: &value{
					observations: 2,
					emptys:       1,
					bools:        2,
				},
			},
			generatorOptions: []GeneratorOption{
				WithSkipUnparseableProperties(false),
			},
			expectedGoType: "map[string]bool",
		},
		{
			name: "object_unparseable_properties_variable_values",
			values: []any{
				map[string]any{
					"key with spaces":         false,
					"another key with spaces": 0,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"key with spaces": {
						observations: 1,
						emptys:       1,
						bools:        1,
					},
					"another key with spaces": {
						observations: 1,
						emptys:       1,
						ints:         1,
					},
				},
				allObjectProperties: &value{
					observations: 2,
					emptys:       2,
					bools:        1,
					ints:         1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithSkipUnparseableProperties(false),
			},
			expectedGoType: "map[string]any",
		},
		{
			name: "object_kebab_case",
			values: []any{
				map[string]any{
					"kebab-case": true,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"kebab-case": {
						observations: 1,
						bools:        1,
					},
				},
				allObjectProperties: &value{
					observations: 1,
					bools:        1,
				},
			},
			expectedGoType: "struct {\nKebabCase bool `json:\"kebab-case\"`\n}",
		},
		{
			name: "string_empty",
			values: []any{
				"",
			},
			expectedValue: &value{
				observations: 1,
				emptys:       1,
				strings:      1,
			},
			expectedGoType: "string",
		},
		{
			name: "string_nonempty",
			values: []any{
				"string",
			},
			expectedValue: &value{
				observations: 1,
				strings:      1,
			},
			expectedGoType: "string",
		},
		{
			name: "string_and_null",
			values: []any{
				"",
				nil,
			},
			expectedValue: &value{
				observations: 2,
				emptys:       1,
				strings:      1,
				nulls:        1,
			},
			expectedGoType: "*string",
		},
		{
			name: "time",
			values: []any{
				"1985-04-12T23:20:50.52Z",
			},
			expectedValue: &value{
				observations: 1,
				strings:      1,
				times:        1,
			},
			expectedGoType: "time.Time",
			expectedImports: map[string]struct{}{
				"time": {},
			},
		},
		{
			name: "time_and_null",
			values: []any{
				"1985-04-12T23:20:50.52Z",
				nil,
			},
			expectedValue: &value{
				observations: 2,
				nulls:        1,
				strings:      1,
				times:        1,
			},
			expectedGoType: "*time.Time",
			expectedImports: map[string]struct{}{
				"time": {},
			},
		},
		{
			name: "time_and_string",
			values: []any{
				"1985-04-12T23:20:50.52Z",
				"",
			},
			expectedValue: &value{
				observations: 2,
				emptys:       1,
				strings:      2,
				times:        1,
			},
			expectedGoType: "string",
		},
		{
			name: "time_and_string_and_null",
			values: []any{
				"1985-04-12T23:20:50.52Z",
				"",
				nil,
			},
			expectedValue: &value{
				observations: 3,
				emptys:       1,
				nulls:        1,
				strings:      2,
				times:        1,
			},
			expectedGoType: "*string",
		},
		{
			name: "custom_export_name_func",
			values: []any{
				map[string]any{
					"gpsAltitude": 0,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"gpsAltitude": {
						observations: 1,
						emptys:       1,
						ints:         1,
					},
				},
				allObjectProperties: &value{
					observations: 1,
					emptys:       1,
					ints:         1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithExtraAbbreviations("GPS"),
			},
			expectedGoType: "struct {\nGPSAltitude int `json:\"gpsAltitude\"`\n}",
		},
		{
			name: "omitempty_always",
			values: []any{
				map[string]any{
					"key": 0,
				},
			},
			expectedValue: &value{
				observations: 1,
				objects:      1,
				objectProperties: map[string]*value{
					"key": {
						observations: 1,
						emptys:       1,
						ints:         1,
					},
				},
				allObjectProperties: &value{
					observations: 1,
					emptys:       1,
					ints:         1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithOmitEmpty(OmitEmptyAlways),
			},
			expectedGoType: "struct {\nKey int `json:\"key,omitempty\"`\n}",
		},
		{
			name: "omitempty_never",
			values: []any{
				map[string]any{
					"key": 0,
				},
				map[string]any{},
			},
			expectedValue: &value{
				observations: 2,
				emptys:       1,
				objects:      2,
				objectProperties: map[string]*value{
					"key": {
						observations: 1,
						emptys:       1,
						ints:         1,
					},
				},
				allObjectProperties: &value{
					observations: 1,
					emptys:       1,
					ints:         1,
				},
			},
			generatorOptions: []GeneratorOption{
				WithOmitEmpty(OmitEmptyNever),
			},
			expectedGoType: "struct {\nKey int `json:\"key\"`\n}",
		},
		{
			name: "omitempty_auto",
			values: []any{
				map[string]any{
					"key1": 0,
					"key2": 0,
				},
				map[string]any{
					"key1": 0,
				},
			},
			expectedValue: &value{
				observations: 2,
				objects:      2,
				objectProperties: map[string]*value{
					"key1": {
						observations: 2,
						emptys:       2,
						ints:         2,
					},
					"key2": {
						observations: 1,
						emptys:       1,
						ints:         1,
					},
				},
				allObjectProperties: &value{
					observations: 3,
					emptys:       3,
					ints:         3,
				},
			},
			generatorOptions: []GeneratorOption{
				WithOmitEmpty(OmitEmptyAuto),
			},
			expectedGoType: "struct {\nKey1 int `json:\"key1\"`\nKey2 int `json:\"key2\"`\n}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			generator := NewGenerator(tc.generatorOptions...)
			for _, value := range tc.values {
				generator.ObserveValue(value)
			}
			assert.Equal(t, tc.expectedValue, generator.value)
			options := &generateOptions{
				exportNameFunc:            generator.exportNameFunc,
				imports:                   make(map[string]struct{}),
				intType:                   generator.intType,
				omitEmptyOption:           generator.omitEmptyOption,
				skipUnparseableProperties: generator.skipUnparseableProperties,
				structTagNames:            generator.structTagNames,
				useJSONNumber:             generator.useJSONNumber,
			}
			goType, _ := generator.value.goType(len(tc.values), options)
			assert.Equal(t, tc.expectedGoType, goType)
			if len(tc.expectedImports) == 0 {
				assert.Equal(t, 0, len(options.imports))
			} else {
				assert.Equal(t, tc.expectedImports, options.imports)
			}
		})
	}
}

func TestObserveJSONGoCode(t *testing.T) {
	for _, tc := range []struct {
		skip              string
		name              string
		json              string
		wantErr           bool
		generatorOptions  []GeneratorOption
		expectedGoCodeStr string
	}{
		{
			name: "error",
			json: "" +
				`"`,
			wantErr: true,
		},
		{
			name: "empty",
			json: "" +
				``,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T any\n",
		},
		{
			name: "bool",
			json: "" +
				`true`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T bool\n",
		},
		{
			name: "int",
			json: "" +
				`0`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T int\n",
		},
		{
			name: "float64",
			json: "" +
				`0.0`,
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
			name: "multiple_tags",
			json: "" +
				`{"intKey":0,"boolKey":true}`,
			generatorOptions: []GeneratorOption{
				WithAddStructTagName("yaml"),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tBoolKey bool `json:\"boolKey\" yaml:\"boolKey\"`\n" +
				"\tIntKey  int  `json:\"intKey\" yaml:\"intKey\"`\n" +
				"}\n",
		},
		{
			name: "multiple_tags_omitempty",
			json: "" +
				`{"intKey":0,"boolKey":true}` +
				`{"intKey":0}`,
			generatorOptions: []GeneratorOption{
				WithStructTagNames([]string{"json", "yaml"}),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tBoolKey bool `json:\"boolKey,omitempty\" yaml:\"boolKey,omitempty\"`\n" +
				"\tIntKey  int  `json:\"intKey\" yaml:\"intKey\"`\n" +
				"}\n",
		},
		{
			name: "empty_component_in_property",
			json: "" +
				`{"int--key":0}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tIntKey int `json:\"int--key\"`\n" +
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
		{
			name: "empty_object",
			json: "" +
				`{}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct{}\n",
		},
		{
			name: "object_and_null",
			json: "" +
				`null` +
				`{}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T *struct{}\n",
		},
		{
			name: "nested_object_always_present_sometimes_null",
			json: "" +
				`{"object":null}` +
				`{"object":{"int":1}}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tObject *struct {\n" +
				"\t\tInt int `json:\"int\"`\n" +
				"\t} `json:\"object\"`\n" +
				"}\n",
		},
		{
			name: "nested_object_always_present_never_null",
			json: "" +
				`{"object":{}}` +
				`{"object":{"int":1}}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tObject struct {\n" +
				"\t\tInt int `json:\"int,omitempty\"`\n" +
				"\t} `json:\"object\"`\n" +
				"}\n",
		},
		{
			name: "nested_object_sometimes_present_never_null",
			json: "" +
				`{}` +
				`{"object":{"int":1}}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tObject *struct {\n" +
				"\t\tInt int `json:\"int\"`\n" +
				"\t} `json:\"object,omitempty\"`\n" +
				"}\n",
		},
		{
			name: "nested_object_sometimes_empty_sometimes_null",
			json: "" +
				`{}` +
				`{"object":null}` +
				`{"object":{}}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tObject *struct{} `json:\"object\"`\n" +
				"}\n",
		},
		{
			name: "nested_empty_object_sometimes_present_never_null",
			json: "" +
				`{}` +
				`{"object":{}}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tObject *struct{} `json:\"object,omitempty\"`\n" +
				"}\n",
		},
		{
			name: "float_and_int_json_number",
			json: "" +
				`{"foo":1}` +
				`{"foo":2.0}`,
			generatorOptions: []GeneratorOption{
				WithUseJSONNumber(true),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"import (\n" +
				"\t\"encoding/json\"\n" +
				")\n" +
				"\n" +
				"type T struct {\n" +
				"\tFoo json.Number `json:\"foo\"`\n" +
				"}\n",
		},
		{
			name: "empty_object_no_go_format",
			json: "" +
				`{}`,
			generatorOptions: []GeneratorOption{
				WithGoFormat(false),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"type T struct{}\n",
		},
		{
			skip: "case fails, needs investigation",
			name: "nested_object_sometimes_present_sometimes_null",
			json: "" +
				`{}` +
				`{"object":null}` +
				`{"object":{"int":1}}`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tObject *struct {\n" +
				"\t\tInt int `json:\"int\"`\n" +
				"\t} `json:\"object\"`\n" +
				"}\n",
		},
		{
			name: "custom_imports_one",
			json: `{}`,
			generatorOptions: []GeneratorOption{
				WithImports("custom_import"),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"import (\n" +
				"\t\"custom_import\"\n" +
				")\n" +
				"\n" +
				"type T struct{}\n",
		},
		{
			name: "custom_imports_multiple",
			json: `{}`,
			generatorOptions: []GeneratorOption{
				WithImports("custom_import_one", "custom_import_two"),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"import (\n" +
				"\t\"custom_import_one\"\n" +
				"\t\"custom_import_two\"\n" +
				")\n" +
				"\n" +
				"type T struct{}\n",
		},
		{
			name: "custom_abbreviations",
			json: `{"my-abbr":true}`,
			generatorOptions: []GeneratorOption{
				WithAbbreviations("ABBR"),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tMyABBR bool `json:\"my-abbr\"`\n" +
				"}\n",
		},
		{
			name: "custom_export_name_func",
			json: `{"myproperty":true}`,
			generatorOptions: []GeneratorOption{
				WithExportNameFunc(strings.ToUpper),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tMYPROPERTY bool `json:\"myproperty\"`\n" +
				"}\n",
		},
		{
			name: "custom_abbreviations",
			json: `{"my-foo":true}`,
			generatorOptions: []GeneratorOption{
				WithAbbreviations("FOO"),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tMyFOO bool `json:\"my-foo\"`\n" +
				"}\n",
		},
		{
			name: "custom_rename",
			json: `{"name":true}`,
			generatorOptions: []GeneratorOption{
				WithRenames(map[string]string{
					"name": "Rename",
				}),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tRename bool `json:\"name\"`\n" +
				"}\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}
			generator := NewGenerator(tc.generatorOptions...)
			err := generator.ObserveJSONReader(bytes.NewBufferString(tc.json))
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			goCode, err := generator.Generate()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedGoCodeStr, string(goCode))
		})
	}
}

func TestObserveYAMLGoCode(t *testing.T) {
	for _, tc := range []struct {
		skip              string
		name              string
		yaml              string
		wantErr           bool
		generatorOptions  []GeneratorOption
		expectedGoCodeStr string
	}{
		{
			name: "empty",
			yaml: "",
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T any\n",
		},
		{
			name:    "error",
			yaml:    "\"",
			wantErr: true,
		},
		{
			name: "bool",
			yaml: "" +
				`true`,
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T bool\n",
		},
		{
			name: "strings",
			yaml: "---\n" +
				"\"a\"\n" +
				"---\n" +
				"\"b\"\n",
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T string\n",
		},
		{
			name: "object",
			yaml: "int: 0\n",
			generatorOptions: []GeneratorOption{
				WithStructTagName("yaml"),
			},
			expectedGoCodeStr: "" +
				"package main\n" +
				"\n" +
				"type T struct {\n" +
				"\tInt int `yaml:\"int\"`\n" +
				"}\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}
			generator := NewGenerator(tc.generatorOptions...)
			err := generator.ObserveYAMLReader(bytes.NewBufferString(tc.yaml))
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			goCode, err := generator.Generate()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedGoCodeStr, string(goCode))
		})
	}
}

func TestObserveJSONFileErrors(t *testing.T) {
	err := NewGenerator().ObserveJSONFile("testdata/notexist.json")
	assert.True(t, errors.Is(err, fs.ErrNotExist))
}

func TestObserveYAMLFileErrors(t *testing.T) {
	err := NewGenerator().ObserveYAMLFile("testdata/notexist.yaml")
	assert.True(t, errors.Is(err, fs.ErrNotExist))
}

func ExampleGenerator_ObserveJSONFile() {
	generator := NewGenerator()
	if err := generator.ObserveJSONFile("testdata/example.json"); err != nil {
		panic(err)
	}
	data, err := generator.Generate()
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(data)

	// Output:
	// package main
	//
	// type T struct {
	// 	Age           int      `json:"age"`
	// 	FavoriteFoods []string `json:"favoriteFoods,omitempty"`
	// 	UserHeightM   float64  `json:"user_height_m"`
	// }
}

func ExampleGenerator_ObserveYAMLFile() {
	generator := NewGenerator(
		WithPackageName("mypackage"),
		WithTypeName("MyType"),
	)
	if err := generator.ObserveYAMLFile("testdata/example.yaml"); err != nil {
		panic(err)
	}
	data, err := generator.Generate()
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(data)

	// Output:
	// package mypackage
	//
	// type MyType struct {
	// 	Nested struct {
	// 		Bar bool    `json:"bar"`
	// 		Foo *string `json:"foo"`
	// 	} `json:"nested"`
	// }
}
