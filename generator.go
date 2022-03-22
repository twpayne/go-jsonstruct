package jsonstruct

// FIXME move substructs to top level

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strings"
)

// An OmitEmptyOption is an option for handling omitempty.
type OmitEmptyOption int

// omitempty options.
const (
	OmitEmptyNever OmitEmptyOption = iota
	OmitEmptyAlways
	OmitEmptyAuto
)

//nolint:gochecknoglobals
var omitEmptyModifier = map[bool]string{
	false: "",
	true:  ",omitempty",
}

// A Generator generates Go types from ObservedValues.
type Generator struct {
	omitEmptyOption           OmitEmptyOption
	fieldNamer                FieldNamer
	skipUnparseableProperties bool
	packageComment            string
	packageName               string
	typeComment               string
	typeName                  string
	structTagNames            []string
	intType                   string
	useJSONNumber             bool
	goFormat                  bool
}

// A GeneratorOption sets an option on a Generator.
type GeneratorOption func(*Generator)

// WithFieldNamer sets the fieldNamer.
func WithFieldNamer(fieldNamer FieldNamer) GeneratorOption {
	return func(g *Generator) {
		g.fieldNamer = fieldNamer
	}
}

// WithGoFormat sets whether the output is should be formatted with go fmt.
func WithGoFormat(goFormat bool) GeneratorOption {
	return func(g *Generator) {
		g.goFormat = goFormat
	}
}

// WithIntType sets the integer type.
func WithIntType(intType string) GeneratorOption {
	return func(g *Generator) {
		g.intType = intType
	}
}

// WithOmitEmpty sets whether each field is tagged with omitempty.
func WithOmitEmpty(omitEmptyOption OmitEmptyOption) GeneratorOption {
	return func(g *Generator) {
		g.omitEmptyOption = omitEmptyOption
	}
}

// WithPackageComment sets the package comment.
func WithPackageComment(packageComment string) GeneratorOption {
	return func(g *Generator) {
		g.packageComment = packageComment
	}
}

// WithPackageName sets the package name.
func WithPackageName(packageName string) GeneratorOption {
	return func(g *Generator) {
		g.packageName = packageName
	}
}

// WithSkipUnparseableProperties sets whether unparseable properties should be
// skipped.
func WithSkipUnparseableProperties(skipUnparseableProperties bool) GeneratorOption {
	return func(g *Generator) {
		g.skipUnparseableProperties = skipUnparseableProperties
	}
}

// WithStructTagName sets the struct tag name.
func WithStructTagName(structTagName string) GeneratorOption {
	return func(g *Generator) {
		g.structTagNames = []string{structTagName}
	}
}

// WithStructTagNames sets the struct tag names.
func WithStructTagNames(structTagNames []string) GeneratorOption {
	return func(g *Generator) {
		g.structTagNames = structTagNames
	}
}

// WithAddStructTagName add struct tag name.
func WithAddStructTagName(structTagName string) GeneratorOption {
	return func(g *Generator) {
		g.structTagNames = append(g.structTagNames, structTagName)
	}
}

// WithTypeComment sets the type comment.
func WithTypeComment(typeComment string) GeneratorOption {
	return func(g *Generator) {
		g.typeComment = typeComment
	}
}

// WithTypeName sets the type name.
func WithTypeName(typeName string) GeneratorOption {
	return func(g *Generator) {
		g.typeName = typeName
	}
}

// WithUseJSONNumber sets whether to use json.Number when both int and float64s
// are observed for the same property.
func WithUseJSONNumber(useJSONNumber bool) GeneratorOption {
	return func(g *Generator) {
		g.useJSONNumber = useJSONNumber
	}
}

// NewGenerator returns a new Generator with options.
func NewGenerator(options ...GeneratorOption) *Generator {
	g := &Generator{
		omitEmptyOption:           OmitEmptyAuto,
		fieldNamer:                defaultFieldNamer,
		skipUnparseableProperties: true,
		packageName:               "main",
		typeName:                  "T",
		structTagNames:            []string{"json"},
		intType:                   "int",
		useJSONNumber:             false,
		goFormat:                  true,
	}
	for _, o := range options {
		o(g)
	}
	return g
}

// GoCode returns the Go source code for o.
func (g *Generator) GoCode(observedValue *ObservedValue) ([]byte, error) {
	buffer := &bytes.Buffer{}
	if g.packageComment != "" {
		fmt.Fprintf(buffer, "// %s\n", g.packageComment)
	}
	fmt.Fprintf(buffer, "package %s\n", g.packageName)
	imports := make(map[string]struct{})
	goType, _ := g.GoType(observedValue, 0, imports)
	if len(imports) > 0 {
		importsSlice := make([]string, 0, len(imports))
		for _import := range imports {
			importsSlice = append(importsSlice, _import)
		}
		sort.Strings(importsSlice)
		fmt.Fprintf(buffer, "import (\n")
		for _, _import := range importsSlice {
			fmt.Fprintf(buffer, "\"%s\"\n", _import)
		}
		fmt.Fprintf(buffer, ")\n")
	}
	if g.typeComment != "" {
		fmt.Fprintf(buffer, "// %s\n", g.typeComment)
	}
	fmt.Fprintf(buffer, "type %s %s\n", g.typeName, goType)
	if !g.goFormat {
		return buffer.Bytes(), nil
	}
	return format.Source(buffer.Bytes())
}

// GoType returns the Go type for o and whether it has been omitted.
func (g *Generator) GoType(o *ObservedValue, observations int, imports map[string]struct{}) (string, bool) {
	// Determine the number of distinct types observed.
	distinctTypes := 0
	if o.Array > 0 {
		distinctTypes++
	}
	if o.Bool > 0 {
		distinctTypes++
	}
	if o.Float64 > 0 {
		distinctTypes++
	}
	if o.Int > 0 {
		distinctTypes++
	}
	if o.Null > 0 {
		distinctTypes++
	}
	if o.Object > 0 {
		distinctTypes++
	}
	if o.String > 0 {
		distinctTypes++
	}

	// Based on the observed distinct types, find the most specific Go type.
	switch {
	case distinctTypes == 1 && o.Array > 0:
		fallthrough
	case distinctTypes == 2 && o.Array > 0 && o.Null > 0:
		elementGoType, _ := g.GoType(o.AllArrayElementValues, 0, imports)
		return "[]" + elementGoType, o.Array+o.Null < observations && o.Empty == 0
	case distinctTypes == 1 && o.Bool > 0:
		return "bool", o.Bool < observations && o.Empty == 0
	case distinctTypes == 2 && o.Bool > 0 && o.Null > 0:
		return "*bool", false
	case distinctTypes == 1 && o.Float64 > 0:
		return "float64", o.Float64 < observations && o.Empty == 0
	case distinctTypes == 2 && o.Float64 > 0 && o.Null > 0:
		return "*float64", false
	case distinctTypes == 1 && o.Int > 0:
		return g.intType, o.Int < observations && o.Empty == 0
	case distinctTypes == 2 && o.Int > 0 && o.Null > 0:
		return "*" + g.intType, false
	case distinctTypes == 2 && o.Float64 > 0 && o.Int > 0:
		omitEmpty := o.Float64+o.Int < observations && o.Empty == 0
		if g.useJSONNumber {
			imports["encoding/json"] = struct{}{}
			return "json.Number", omitEmpty
		}
		return "float64", omitEmpty
	case distinctTypes == 3 && o.Float64 > 0 && o.Int > 0 && o.Null > 0:
		if g.useJSONNumber {
			imports["encoding/json"] = struct{}{}
			return "*json.Number", false
		}
		return "*float64", false
	case distinctTypes == 1 && o.Object > 0:
		fallthrough
	case distinctTypes == 2 && o.Object > 0 && o.Null > 0:
		if len(o.ObjectPropertyValue) == 0 {
			switch {
			case observations == 0 && o.Null == 0:
				return "struct{}", false
			case o.Null > 0:
				return "*struct{}", false
			case o.Object == observations:
				return "struct{}", false
			default:
				return "*struct{}", o.Object < observations
			}
		}
		hasUnparseableProperties := false
		for k := range o.ObjectPropertyValue {
			if strings.ContainsRune(k, ' ') {
				hasUnparseableProperties = true
				break
			}
		}
		if hasUnparseableProperties && !g.skipUnparseableProperties {
			valueGoType, _ := g.GoType(o.AllObjectPropertyValues, 0, imports)
			return "map[string]" + valueGoType, o.Object+o.Null < observations
		}
		b := &bytes.Buffer{}
		properties := make([]string, 0, len(o.ObjectPropertyValue))
		for k := range o.ObjectPropertyValue {
			properties = append(properties, k)
		}
		sort.Strings(properties)
		fmt.Fprintf(b, "struct {\n")
		var unparseableProperties []string
		for _, k := range properties {
			if isUnparseableProperty(k) {
				unparseableProperties = append(unparseableProperties, k)
				continue
			}
			goType, observedEmpty := g.GoType(o.ObjectPropertyValue[k], o.Object, imports)
			var omitEmpty bool
			switch {
			case g.omitEmptyOption == OmitEmptyNever:
				omitEmpty = false
			case g.omitEmptyOption == OmitEmptyAlways:
				omitEmpty = true
			case g.omitEmptyOption == OmitEmptyAuto:
				omitEmpty = observedEmpty
			}

			tag := "`"

			for i, v := range g.structTagNames {
				if i != 0 {
					tag += " "
				}

				tag += fmt.Sprintf("%s:\"%s%s\"", v, k, omitEmptyModifier[omitEmpty])
			}

			tag += "`"

			fmt.Fprintf(b, "%s %s %s\n", g.fieldNamer.FieldName(k), goType, tag)
		}
		for _, k := range unparseableProperties {
			fmt.Fprintf(b, "// %q cannot be unmarshalled into a struct field by encoding/json.\n", k)
		}
		fmt.Fprintf(b, "}")
		switch {
		case observations == 0:
			return b.String(), false
		case o.Object == observations:
			return b.String(), false
		case o.Object < observations && o.Null == 0:
			return "*" + b.String(), true
		default:
			return "*" + b.String(), o.Object+o.Null < observations
		}
	case distinctTypes == 1 && o.String > 0 && o.Time == o.String:
		imports["time"] = struct{}{}
		return "time.Time", o.Time < observations
	case distinctTypes == 1 && o.String > 0:
		return "string", o.String < observations && o.Empty == 0
	case distinctTypes == 2 && o.String > 0 && o.Null > 0 && o.Time == o.String:
		imports["time"] = struct{}{}
		return "*time.Time", false
	case distinctTypes == 2 && o.String > 0 && o.Null > 0:
		return "*string", false
	default:
		return "interface{}", o.Array+o.Bool+o.Float64+o.Int+o.Null+o.Object+o.String < observations
	}
}

// isUnparseableProperty returns true if key cannot be parsed by encoding/json.
func isUnparseableProperty(key string) bool {
	return strings.ContainsAny(key, ` ",`)
}
