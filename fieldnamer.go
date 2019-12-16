package jsonstruct

import (
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
)

//nolint:gochecknoglobals
var (
	WellKnownAbbreviations = map[string]bool{
		"API":  true,
		"DB":   true,
		"HTTP": true,
		"ID":   true,
		"JSON": true,
		"SQL":  true,
		"URI":  true,
		"URL":  true,
		"XML":  true,
	}

	defaultFieldNamer = &AbbreviationHandlingFieldNamer{
		Abbreviations: WellKnownAbbreviations,
	}
)

// A FieldNamer generates a Go field name from a JSON property.
type FieldNamer interface {
	FieldName(property string) string
}

// An AbbreviationHandlingFieldNamer generates Go field names from JSON
// properties while keeping abbreviations uppercased.
type AbbreviationHandlingFieldNamer struct {
	Abbreviations map[string]bool
}

// FieldName implements FieldNamer.FieldName.
func (a *AbbreviationHandlingFieldNamer) FieldName(property string) string {
	components := SplitComponents(property)
	for i, component := range components {
		switch {
		case component == "":
			// do nothing
		case a.Abbreviations[strings.ToUpper(component)]:
			components[i] = strings.ToUpper(component)
		default:
			runes := []rune(component)
			runes[0] = unicode.ToUpper(runes[0])
			components[i] = string(runes)
		}
	}
	return strings.Join(components, "")
}

// SplitComponents splits name into components. name may be kebab case, snake
// case, or camel case.
func SplitComponents(name string) []string {
	switch {
	case strings.ContainsRune(name, '-'):
		return strings.Split(name, "-")
	case strings.ContainsRune(name, '_'):
		return strings.Split(name, "_")
	default:
		return camelcase.Split(name)
	}
}
