package jsonstruct

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"maps"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// An ExportNameFunc returns the exported name for a property.
type ExportNameFunc func(string) string

// An OmitEmptyTagsType sets how to handle ,omitempty tags.
type OmitEmptyTagsType int

// OmitEmptyTags values.
const (
	OmitEmptyTagsNever OmitEmptyTagsType = iota
	OmitEmptyTagsAlways
	OmitEmptyTagsAuto
)

// A Generator generates Go types from observed values.
type Generator struct {
	abbreviations            map[string]bool
	exportNameFunc           ExportNameFunc
	exportRenames            map[string]string
	fileHeader               string
	goFormat                 bool
	imports                  map[string]struct{}
	intType                  string
	omitEmptyTags            OmitEmptyTagsType
	packageComment           string
	packageName              string
	skipUnparsableProperties bool
	stringTags               bool
	structTagNames           []string
	typeComment              string
	typeName                 string
	useJSONNumber            bool
	value                    *value
}

// A GeneratorOption sets an option on a Generator.
type GeneratorOption func(*Generator)

// WithAbbreviations sets the abbreviations.
func WithAbbreviations(abbreviations ...string) GeneratorOption {
	return func(g *Generator) {
		g.abbreviations = make(map[string]bool)
		for _, abbreviation := range abbreviations {
			g.abbreviations[abbreviation] = true
		}
	}
}

// WithExportNameFunc sets the export name function.
func WithExportNameFunc(exportNameFunc ExportNameFunc) GeneratorOption {
	return func(g *Generator) {
		g.exportNameFunc = exportNameFunc
	}
}

// WithExtraAbbreviations adds abbreviations.
func WithExtraAbbreviations(abbreviations ...string) GeneratorOption {
	return func(g *Generator) {
		for _, abbreviation := range abbreviations {
			g.abbreviations[abbreviation] = true
		}
	}
}

// WithFileHeader sets the file header.
func WithFileHeader(fileHeader string) GeneratorOption {
	return func(g *Generator) {
		g.fileHeader = fileHeader
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

// WithOmitEmptyTags sets whether ",omitempty" tags should be used.
func WithOmitEmptyTags(omitEmptyTags OmitEmptyTagsType) GeneratorOption {
	return func(g *Generator) {
		g.omitEmptyTags = omitEmptyTags
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

// WithRenames sets the renames.
func WithRenames(renames map[string]string) GeneratorOption {
	return func(g *Generator) {
		for name, rename := range renames {
			g.exportRenames[name] = rename
		}
	}
}

// WithSkipUnparsableProperties sets whether unparsable properties should be
// skipped.
func WithSkipUnparsableProperties(skipUnparsableProperties bool) GeneratorOption {
	return func(g *Generator) {
		g.skipUnparsableProperties = skipUnparsableProperties
	}
}

// WithStringTags sets whether ",string" tags should be used.
func WithStringTags(stringTags bool) GeneratorOption {
	return func(g *Generator) {
		g.stringTags = stringTags
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

// WithAddStructTagName adds a struct tag name.
func WithAddStructTagName(structTagName string) GeneratorOption {
	return func(g *Generator) {
		g.structTagNames = append(g.structTagNames, structTagName)
	}
}

// WithImports adds custom package imports.
func WithImports(imports ...string) GeneratorOption {
	return func(g *Generator) {
		for _, v := range imports {
			g.imports[v] = struct{}{}
		}
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
		abbreviations:            maps.Clone(defaultAbbreviations),
		exportRenames:            make(map[string]string),
		goFormat:                 true,
		imports:                  make(map[string]struct{}),
		intType:                  "int",
		omitEmptyTags:            OmitEmptyTagsAuto,
		packageName:              "main",
		skipUnparsableProperties: true,
		structTagNames:           []string{"json"},
		typeName:                 "T",
		useJSONNumber:            false,
		value:                    &value{},
	}
	g.exportNameFunc = func(name string) string {
		if rename, ok := g.exportRenames[name]; ok {
			return rename
		}
		return DefaultExportNameFunc(name, g.abbreviations)
	}
	for _, option := range options {
		option(g)
	}
	return g
}

// Generate returns the Go source code for the observed values.
func (g *Generator) Generate() ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.Grow(65536)
	if g.fileHeader != "" {
		fmt.Fprintf(buffer, "%s\n\n", g.fileHeader)
	}
	if g.packageComment != "" {
		fmt.Fprintf(buffer, "// %s\n", g.packageComment)
	}
	fmt.Fprintf(buffer, "package %s\n", g.packageName)
	imports := maps.Clone(g.imports)
	goType := g.value.goType(0, &generateOptions{
		exportNameFunc:           g.exportNameFunc,
		imports:                  imports,
		intType:                  g.intType,
		omitEmptyTags:            g.omitEmptyTags,
		skipUnparsableProperties: g.skipUnparsableProperties,
		stringTags:               g.stringTags,
		structTagNames:           g.structTagNames,
		useJSONNumber:            g.useJSONNumber,
	})
	if len(imports) > 0 {
		importsSlice := sortedKeys(imports)
		fmt.Fprintf(buffer, "import (\n")
		for _, _import := range importsSlice {
			fmt.Fprintf(buffer, "\"%s\"\n", _import)
		}
		fmt.Fprintf(buffer, ")\n")
	}
	if g.typeComment != "" {
		fmt.Fprintf(buffer, "// %s\n", g.typeComment)
	}
	fmt.Fprintf(buffer, "type %s %s\n", g.typeName, goType.typeStr)
	if !g.goFormat {
		return buffer.Bytes(), nil
	}
	return format.Source(buffer.Bytes())
}

// ObserveValue observes value.
func (g *Generator) ObserveValue(value any) {
	g.value = g.value.observe(value)
}

// ObserveJSONReader observes JSON values from r.
func (g *Generator) ObserveJSONReader(r io.Reader) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	for {
		var value any
		err := decoder.Decode(&value)
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		default:
			g.ObserveValue(value)
		}
	}
}

// ObserveJSONFile observes JSON values from filename.
func (g *Generator) ObserveJSONFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return g.ObserveJSONReader(file)
}

// ObserveYAMLReader observes YAML values from r.
func (g *Generator) ObserveYAMLReader(r io.Reader) error {
	decoder := yaml.NewDecoder(r)
	for {
		var value any
		err := decoder.Decode(&value)
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		default:
			g.ObserveValue(value)
		}
	}
}

// ObserveYAMLFile observes YAML values from filename.
func (g *Generator) ObserveYAMLFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return g.ObserveYAMLReader(file)
}

// isUnparsableProperty returns true if key cannot be parsed by encoding/json.
func isUnparsableProperty(key string) bool {
	return strings.ContainsAny(key, ` ",`)
}
