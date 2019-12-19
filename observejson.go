package jsonstruct

import (
	"encoding/json"
	"errors"
	"io"
)

// ObserveJSON returns all JSON values observed in r.
func ObserveJSON(r io.Reader) (*ObservedValue, error) {
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
