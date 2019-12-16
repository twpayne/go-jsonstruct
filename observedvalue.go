package jsonstruct

import (
	"encoding/json"
	"errors"
	"io"
	"time"
)

// FIXME extract sub-structs

// An ObservedValue is an observed value.
type ObservedValue struct {
	Observations            int
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
		if o.AllArrayElementValues == nil {
			o.AllArrayElementValues = &ObservedValue{}
		}
		for _, e := range value {
			o.AllArrayElementValues = o.AllArrayElementValues.Merge(e)
		}
	case bool:
		o.Bool++
	case float64:
		o.Float64++
	case int:
		o.Int++
	case nil:
		o.Null++
	case map[string]interface{}:
		o.Object++
		if o.ObjectPropertyValue == nil {
			o.ObjectPropertyValue = make(map[string]*ObservedValue)
		}
		for k, v := range value {
			o.AllObjectPropertyValues = o.AllObjectPropertyValues.Merge(v)
			o.ObjectPropertyValue[k] = o.ObjectPropertyValue[k].Merge(v)
		}
	case string:
		o.String++
		if _, err := time.Parse(time.RFC3339Nano, value); err == nil {
			o.Time++
		}
	case json.Number:
		if _, err := value.Int64(); err == nil {
			o.Int++
		} else {
			o.Float64++
		}
	}
	return o
}

// Observe returns all values observed in r.
func Observe(r io.Reader) (*ObservedValue, error) {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	observedValue := &ObservedValue{}
	for {
		var value interface{}
		err := decoder.Decode(&value)
		switch {
		case errors.Is(err, io.EOF):
			return observedValue, nil
		case err != nil:
			return nil, err
		default:
			observedValue = observedValue.Merge(value)
		}
	}
}
