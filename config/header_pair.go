package config

import (
	"encoding/json"
	"fmt"

	"github.com/ironsmile/nedomi/utils"
)

// HeaderPairs is map between string and string or string and []string through StringSlice
type HeaderPairs map[string]StringSlice

func interfacesToStrings(input []interface{}) ([]string, error) {
	var output = make([]string, len(input))
	for index, inter := range input {
		str, ok := inter.(string)
		if !ok { // :(
			return nil, fmt.Errorf("the `%+v` is not string", inter)
		}
		output[index] = str
	}

	return output, nil
}

// Copy makes a deep copy of the HeaderPairs
func (h HeaderPairs) Copy() HeaderPairs {
	var res = make(HeaderPairs)
	for key, value := range h {
		res[key] = utils.CopyStringSlice(value)
	}
	return res
}

// StringSlice is a slice of string with unmarshalling
// powers of turning a string into a slice of strings
type StringSlice []string

// UnmarshalJSON unmarshalles StringSlice and checks that it's valid
func (s *StringSlice) UnmarshalJSON(buf []byte) error {
	var tmp interface{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}

	var result []string
	switch v := tmp.(type) {
	case string:
		result = []string{v}
	case []string: // this doesn't actually happen
		result = v
	case []interface{}: // the actuall []string, hopefully
		var strs, err = interfacesToStrings(v)
		if err != nil {
			return err
		}
		result = strs
	default:
		return fmt.Errorf("StringSlice is neither string nor slice of strings but %+v", tmp)
	}
	*s = result
	return nil
}
