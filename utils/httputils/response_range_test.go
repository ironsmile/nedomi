package httputils

import (
	"reflect"
	"testing"
)

type respRangeTest struct {
	r        string
	expRange *HTTPContentRange
	expErr   bool
}

var respRangeTests = []respRangeTest{
	{r: "wrong", expErr: true},
	{r: "bytes", expErr: true},
	{r: "bytes ", expErr: true},
	{r: "1-2/11", expErr: true},
	{r: "bytes=1-2/11", expErr: true},
	{r: "bytes a-1/11", expErr: true},
	{r: "bytes 2-a/11", expErr: true},
	{r: "bytes -1-2/11", expErr: true},
	{r: "bytes 1--2/11", expErr: true},
	{r: "bytes 1a-2/11", expErr: true},
	{r: "bytes 1-2a/11", expErr: true},
	{r: "bytes 2-1/11", expErr: true},
	{r: "bytes 2-/11", expErr: true},
	{r: "bytes -1/11", expErr: true},
	{r: "bytes 10-11/11", expErr: true},
	{r: "bytes 0-0/0", expErr: true},
	{r: "bytes */11", expErr: true},
	{r: "bytes 1-2/11a", expErr: true},
	{r: "bytes 1-2/*", expErr: true},

	{r: "", expRange: nil},
	{r: "bytes 0-0/1", expRange: &HTTPContentRange{0, 1, 1}},
	{r: "bytes 0-4/11", expRange: &HTTPContentRange{0, 5, 11}},
	{r: "bytes 2-10/12", expRange: &HTTPContentRange{2, 9, 12}},
	{r: "bytes 1-5/13", expRange: &HTTPContentRange{1, 5, 13}},
	{r: "bytes 13-13/14", expRange: &HTTPContentRange{13, 1, 14}},
}

func TestResponseContentRangeParsing(t *testing.T) {
	t.Parallel()
	for _, test := range respRangeTests {
		res, err := ParseRespContentRange(test.r)

		if err != nil && !test.expErr {
			t.Errorf("Received an unexpected error for test %q: %s", test.r, err)
		}
		if err == nil && test.expErr {
			t.Errorf("Expected to receive an error for test %q", test.r)
		}
		if !reflect.DeepEqual(res, test.expRange) {
			t.Errorf("The received range for test %q '%#v' differ from the expected '%#v'", test.r, res, test.expRange)
		}
	}
}
