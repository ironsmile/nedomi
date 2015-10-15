package config

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestHeaderPairsParsing(t *testing.T) {
	var tests = []struct {
		key      string
		expected []string
	}{
		{key: "1", expected: []string{"2"}},
		{key: "3", expected: []string{"4", "5"}},
		{key: "4", expected: nil},
	}

	var input = `{ "add_headers": {
		"1": "2",
		"3" : [ "4", "5"]
	}}`

	var s struct {
		AddHeaders HeaderPairs `json:"add_headers"`
	}
	if err := json.Unmarshal([]byte(input), &s); err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		got := ([]string)(s.AddHeaders[test.key]) // reflect reasons
		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("for '%s': expected '%+v' got '%+v'", test.key, test.expected, got)
		}
	}
}

func TestStringSliceNotString(t *testing.T) {
	var inputs = []string{
		`{"values" : ["2", "3", 4]}`, // 4 is not a string
		`{"values" : 5}`,             // 5 is not a string either
	}

	var s struct {
		Values StringSlice `json:"values"`
	}
	for _, input := range inputs {
		if err := json.Unmarshal([]byte(input), &s); err == nil {
			t.Errorf("expected error for input '%s' but got values : %+v",
				input, s.Values)
		} else {
			t.Log(err)
		}
	}
}
