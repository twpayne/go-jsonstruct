package jsonstruct

import (
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
)

//nolint:gochecknoglobals
var (
	defaultAbbreviations = map[string]bool{
		"ACL":   true,
		"API":   true,
		"DB":    true,
		"HTTP":  true,
		"HTTPS": true,
		"ID":    true,
		"JSON":  true,
		"OS":    true,
		"SQL":   true,
		"SSH":   true,
		"URI":   true,
		"URL":   true,
		"UUID":  true,
		"XML":   true,
		"YAML":  true,
	}
)

// DefaultExportNameFunc returns the exported name for name.
func DefaultExportNameFunc(name string, abbreviations map[string]bool) string {
	components := SplitComponents(name)
	for i, component := range components {
		switch {
		case component == "":
			// do nothing
		case abbreviations[strings.ToUpper(component)]:
			components[i] = strings.ToUpper(component)
		case component == strings.ToUpper(component):
			runes := []rune(component)
			components[i] = string(runes[0]) + strings.ToLower(string(runes[1:]))
		default:
			runes := []rune(component)
			runes[0] = unicode.ToUpper(runes[0])
			components[i] = string(runes)
		}
	}
	runes := []rune(strings.Join(components, ""))
	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			runes[i] = '_'
		}
	}
	exportName := string(runes)
	if !unicode.IsLetter(runes[0]) && runes[0] != '_' {
		exportName = "_" + exportName
	}
	return exportName
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
