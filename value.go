package jsonstruct

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/structtag"
)

// An value describes an observed value.
type value struct {
	observations        int
	empties             int
	zeros               int
	arrays              int
	bools               int
	boolStrings         int
	float64s            int
	float64Strings      int
	ints                int
	intStrings          int
	nulls               int
	objects             int
	strings             int
	times               int // time.Time is an implicit more specific type than string.
	arrayElements       *value
	allObjectProperties *value
	objectProperties    map[string]*value
}

type generateOptions struct {
	exportNameFunc           ExportNameFunc
	imports                  map[string]struct{}
	intType                  string
	omitEmptyTags            OmitEmptyTagsType
	omitZeroTags             OmitZeroTagsType
	skipUnparsableProperties bool
	stringTags               bool
	structTagNames           []string
	useJSONNumber            bool
}

type goType struct {
	typeStr   string
	omitEmpty bool
	omitZero  bool
	stringTag bool
}

// observe merges a into v.
func (v *value) observe(a any) *value {
	if v == nil {
		v = &value{}
	}
	v.observations++
	switch a := a.(type) {
	case []any:
		v.arrays++
		if len(a) == 0 {
			v.empties++
		}
		if v.arrayElements == nil {
			v.arrayElements = &value{}
		}
		for _, e := range a {
			v.arrayElements = v.arrayElements.observe(e)
		}
	case bool:
		v.bools++
		if !a {
			v.empties++
			v.zeros++
		}
	case float64:
		v.float64s++
		if a == 0 {
			v.empties++
			v.zeros++
		}
	case int:
		v.ints++
		if a == 0 {
			v.empties++
			v.zeros++
		}
	case nil:
		v.nulls++
		v.zeros++
	case map[string]any:
		v.objects++
		if len(a) == 0 {
			v.empties++
		}
		if v.objectProperties == nil {
			v.objectProperties = make(map[string]*value)
		}
		for property, value := range a {
			v.allObjectProperties = v.allObjectProperties.observe(value)
			v.objectProperties[property] = v.objectProperties[property].observe(value)
		}
	case string:
		if a == "" {
			v.empties++
			v.zeros++
		}
		if err := json.Unmarshal([]byte(a), new(bool)); err == nil {
			v.boolStrings++
		} else if err := json.Unmarshal([]byte(a), new(int)); err == nil {
			v.float64Strings++
			v.intStrings++
		} else if err := json.Unmarshal([]byte(a), new(float64)); err == nil {
			v.float64Strings++
		}
		if v.times == v.strings {
			if t, err := time.Parse(time.RFC3339Nano, a); err == nil {
				v.times++
				if t.IsZero() {
					v.zeros++
				}
			}
		}
		v.strings++
	case json.Number:
		if i, err := a.Int64(); err == nil {
			v.ints++
			if i == 0 {
				v.zeros++
			}
		} else {
			v.float64s++
			if f, err := a.Float64(); err == nil && f == 0 {
				v.zeros++
			}
		}
	}
	return v
}

// goType returns the Go type of v.
func (v *value) goType(observations int, options *generateOptions) goType {
	// Determine the number of distinct types observed.
	distinctTypes := 0
	if v.arrays > 0 {
		distinctTypes++
	}
	if v.bools > 0 {
		distinctTypes++
	}
	if v.float64s > 0 {
		distinctTypes++
	}
	if v.ints > 0 {
		distinctTypes++
	}
	if v.nulls > 0 {
		distinctTypes++
	}
	if v.objects > 0 {
		distinctTypes++
	}
	if v.strings > 0 {
		distinctTypes++
	}

	// Based on the observed distinct types, find the most specific Go type.
	switch {
	case distinctTypes == 1 && v.arrays > 0:
		fallthrough
	case distinctTypes == 2 && v.arrays > 0 && v.nulls > 0:
		elementGoType := v.arrayElements.goType(0, options)
		return goType{
			typeStr:   "[]" + elementGoType.typeStr,
			omitEmpty: v.arrays+v.nulls < observations && v.empties == 0,
		}
	case distinctTypes == 1 && v.bools > 0:
		return goType{
			typeStr:   "bool",
			omitEmpty: v.bools < observations && v.empties == 0,
			omitZero:  v.zeros == 0,
		}
	case distinctTypes == 2 && v.bools > 0 && v.nulls > 0:
		return goType{
			typeStr: "*bool",
		}
	case distinctTypes == 1 && v.float64s > 0:
		return goType{
			typeStr:   "float64",
			omitEmpty: v.float64s < observations && v.empties == 0,
			omitZero:  v.zeros == 0,
		}
	case distinctTypes == 2 && v.float64s > 0 && v.nulls > 0:
		return goType{
			typeStr: "*float64",
		}
	case distinctTypes == 1 && v.ints > 0:
		return goType{
			typeStr:   options.intType,
			omitEmpty: v.ints < observations && v.empties == 0,
			omitZero:  v.zeros == 0,
		}
	case distinctTypes == 2 && v.ints > 0 && v.nulls > 0:
		return goType{
			typeStr: "*" + options.intType,
		}
	case distinctTypes == 2 && v.float64s > 0 && v.ints > 0:
		omitEmpty := v.float64s+v.ints < observations && v.empties == 0
		if options.useJSONNumber {
			options.imports["encoding/json"] = struct{}{}
			return goType{
				typeStr:   "json.Number",
				omitEmpty: omitEmpty,
				omitZero:  v.zeros == 0,
			}
		}
		return goType{
			typeStr:   "float64",
			omitEmpty: omitEmpty,
			omitZero:  v.zeros == 0,
		}
	case distinctTypes == 3 && v.float64s > 0 && v.ints > 0 && v.nulls > 0:
		if options.useJSONNumber {
			options.imports["encoding/json"] = struct{}{}
			return goType{
				typeStr:  "*json.Number",
				omitZero: v.zeros == 0,
			}
		}
		return goType{
			typeStr:  "*float64",
			omitZero: v.zeros == 0,
		}
	case distinctTypes == 1 && v.objects > 0:
		fallthrough
	case distinctTypes == 2 && v.objects > 0 && v.nulls > 0:
		if len(v.objectProperties) == 0 {
			switch {
			case observations == 0 && v.nulls == 0:
				return goType{
					typeStr: "struct{}",
				}
			case v.nulls > 0:
				return goType{
					typeStr: "*struct{}",
				}
			case v.objects == observations:
				return goType{
					typeStr: "struct{}",
				}
			default:
				return goType{
					typeStr:   "*struct{}",
					omitEmpty: v.objects < observations,
				}
			}
		}
		hasUnparsableProperties := false
		for k := range v.objectProperties {
			if strings.ContainsRune(k, ' ') {
				hasUnparsableProperties = true
				break
			}
		}
		if hasUnparsableProperties && !options.skipUnparsableProperties {
			valueGoType := v.allObjectProperties.goType(0, options)
			return goType{
				typeStr:   "map[string]" + valueGoType.typeStr,
				omitEmpty: v.objects+v.nulls < observations,
			}
		}
		b := &bytes.Buffer{}
		properties := sortedKeys(v.objectProperties)
		fmt.Fprintf(b, "struct {\n")
		var unparsableProperties []string
		for _, property := range properties {
			if isUnparsableProperty(property) {
				unparsableProperties = append(unparsableProperties, property)
				continue
			}
			goType := v.objectProperties[property].goType(v.objects, options)
			var omitEmpty bool
			switch {
			case options.omitEmptyTags == OmitEmptyTagsNever:
				omitEmpty = false
			case options.omitEmptyTags == OmitEmptyTagsAlways:
				omitEmpty = true
			case options.omitEmptyTags == OmitEmptyTagsAuto:
				omitEmpty = goType.omitEmpty
			}
			var omitZero bool
			switch options.omitZeroTags {
			case OmitZeroTagsNever:
				omitZero = false
			case OmitZeroTagsAlways:
				omitZero = true
			case OmitZeroTagsAuto:
				omitZero = goType.omitZero
			}

			tags, _ := structtag.Parse("")
			var structTagOptions []string
			if omitEmpty {
				structTagOptions = append(structTagOptions, "omitempty")
			}
			if omitZero {
				structTagOptions = append(structTagOptions, "omitzero")
			}
			if goType.stringTag {
				structTagOptions = append(structTagOptions, "string")
			}
			for _, structTagName := range options.structTagNames {
				tag := &structtag.Tag{
					Key:     structTagName,
					Name:    property,
					Options: structTagOptions,
				}
				_ = tags.Set(tag)
			}

			fmt.Fprintf(b, "%s %s `%s`\n", options.exportNameFunc(property), goType.typeStr, tags)
		}
		for _, property := range unparsableProperties {
			fmt.Fprintf(b, "// %q cannot be unmarshalled into a struct field by encoding/json.\n", property)
		}
		fmt.Fprintf(b, "}")
		switch {
		case observations == 0:
			return goType{
				typeStr: b.String(),
			}
		case v.objects == observations:
			return goType{
				typeStr: b.String(),
			}
		case v.objects < observations && v.nulls == 0:
			return goType{
				typeStr:   "*" + b.String(),
				omitEmpty: true,
				omitZero:  v.zeros == 0,
			}
		default:
			return goType{
				typeStr:   "*" + b.String(),
				omitEmpty: v.objects+v.nulls < observations,
				omitZero:  v.zeros == 0,
			}
		}
	case distinctTypes == 1 && v.strings > 0 && v.times == v.strings:
		options.imports["time"] = struct{}{}
		return goType{
			typeStr:   "time.Time",
			omitEmpty: v.times < observations,
			omitZero:  v.zeros == 0,
		}
	case distinctTypes == 1 && v.strings > 0:
		switch {
		case options.stringTags && v.strings == v.boolStrings:
			return goType{
				typeStr:   "bool",
				stringTag: true,
				omitEmpty: v.boolStrings < v.observations,
				omitZero:  v.zeros == 0,
			}
		case options.stringTags && v.strings == v.intStrings:
			return goType{
				typeStr:   options.intType,
				stringTag: true,
				omitEmpty: v.intStrings < v.strings,
				omitZero:  v.zeros == 0,
			}
		case options.stringTags && v.strings == v.float64Strings:
			return goType{
				typeStr:   "float64",
				stringTag: true,
				omitEmpty: v.float64Strings < v.strings,
				omitZero:  v.zeros == 0,
			}
		default:
			return goType{
				typeStr:   "string",
				omitEmpty: v.strings < observations && v.empties == 0,
				omitZero:  v.zeros == 0,
			}
		}
	case distinctTypes == 2 && v.strings > 0 && v.nulls > 0 && v.times == v.strings:
		options.imports["time"] = struct{}{}
		return goType{
			typeStr: "*time.Time",
		}
	case distinctTypes == 2 && v.strings > 0 && v.nulls > 0:
		return goType{
			typeStr: "*string",
		}
	default:
		return goType{
			typeStr:   "any",
			omitEmpty: v.arrays+v.bools+v.float64s+v.ints+v.nulls+v.objects+v.strings < observations,
		}
	}
}
