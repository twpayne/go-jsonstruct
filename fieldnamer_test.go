package jsonstruct

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFieldNamer(t *testing.T) {
	for name, expected := range map[string]string{
		"id":         "ID",
		"foo":        "Foo",
		"foo_bar":    "FooBar",
		"fooBar":     "FooBar",
		"FOO_BAR":    "FooBar",
		"FOO_BAR_ID": "FooBarID",
	} {
		assert.Equal(t, expected, defaultFieldNamer.FieldName(name))
	}
}
