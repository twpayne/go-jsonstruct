package jsonstruct

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDefaultExportNameFunc(t *testing.T) {
	expected := map[string]string{
		"FOO_BAR":          "FooBar",
		"FOO_BAR_ID":       "FooBarID",
		"LIST_OF_OSES":     "ListOfOSes",
		"foo":              "Foo",
		"fooBar":           "FooBar",
		"foo_bar":          "FooBar",
		"https_urls":       "HTTPSURLs",
		"id":               "ID",
		"ids":              "IDs",
		"urls_to_download": "URLsToDownload",
		"user_acls":        "UserACLs",
		"123":              "_123",
		"A|B":              "A_B",
	}
	for _, name := range sortedKeys(expected) {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, expected[name], DefaultExportNameFunc(name, defaultAbbreviations))
		})
	}
}
