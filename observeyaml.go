package jsonstruct

import (
	"errors"
	"io"

	"gopkg.in/yaml.v3"
)

// ObserveYAML returns all YAML values observed in r.
func ObserveYAML(r io.Reader) (*ObservedValue, error) {
	decoder := yaml.NewDecoder(r)
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
