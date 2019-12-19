package jsonstruct

import (
	"encoding/json"
	"time"
)

// FIXME extract sub-structs

// An ObservedValue is an observed value.
type ObservedValue struct {
	Observations            int
	Empty                   int
	Array                   int
	Bool                    int
	Float64                 int
	Int                     int
	Null                    int
	Object                  int
	String                  int
	Time                    int // time.Time is an implicit more specific type than string.
	AllArrayElementValues   *ObservedValue
	AllObjectPropertyValues *ObservedValue
	ObjectPropertyValue     map[string]*ObservedValue
}

// Merge merges value into o.
func (o *ObservedValue) Merge(value interface{}) *ObservedValue {
	if o == nil {
		o = &ObservedValue{}
	}
	o.Observations++
	switch value := value.(type) {
	case []interface{}:
		o.Array++
		if len(value) == 0 {
			o.Empty++
		}
		if o.AllArrayElementValues == nil {
			o.AllArrayElementValues = &ObservedValue{}
		}
		for _, e := range value {
			o.AllArrayElementValues = o.AllArrayElementValues.Merge(e)
		}
	case bool:
		o.Bool++
		if !value {
			o.Empty++
		}
	case float64:
		o.Float64++
		if value == 0 {
			o.Empty++
		}
	case int:
		o.Int++
		if value == 0 {
			o.Empty++
		}
	case nil:
		o.Null++
	case map[string]interface{}:
		o.Object++
		if len(value) == 0 {
			o.Empty++
		}
		if o.ObjectPropertyValue == nil {
			o.ObjectPropertyValue = make(map[string]*ObservedValue)
		}
		for k, v := range value {
			o.AllObjectPropertyValues = o.AllObjectPropertyValues.Merge(v)
			o.ObjectPropertyValue[k] = o.ObjectPropertyValue[k].Merge(v)
		}
	case string:
		if value == "" {
			o.Empty++
		}
		if o.Time == o.String {
			if _, err := time.Parse(time.RFC3339Nano, value); err == nil {
				o.Time++
			}
		}
		o.String++
	case json.Number:
		if _, err := value.Int64(); err == nil {
			o.Int++
		} else {
			o.Float64++
		}
	}
	return o
}
