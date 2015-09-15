package cache

import (
	"fmt"
	"reflect"
	"testing"
)

type reqRangeTest struct {
	r         string
	size      uint64
	expErr    bool
	expRanges []httpRange
}

func (t reqRangeTest) String() string {
	return fmt.Sprintf("(r:%q, size:%d)", t.r, t.size)
}

var reqRangeTests = []reqRangeTest{
	{r: "wrong", expErr: true},
	{r: "bytes", expErr: true},
	{r: "bytes=", expErr: true},
	{r: "1-2", size: 11, expErr: true},
	{r: "bytes=a-1", size: 11, expErr: true},
	{r: "bytes=2-a", size: 11, expErr: true},
	{r: "bytes=-1-2", size: 11, expErr: true},
	{r: "bytes=1--2", size: 11, expErr: true},
	{r: "bytes=1a-2", size: 11, expErr: true},
	{r: "bytes=1-2a", size: 11, expErr: true},
	{r: "bytes=2-1", size: 11, expErr: true},
	{r: "bytes=12-13", size: 11, expErr: true},
	{r: "bytes=0-0", size: 0, expErr: true},
	{r: "bytes=0-", size: 0, expErr: true},
	{r: "bytes=-0", size: 1, expErr: true},

	{r: "bytes=0-0", size: 1, expRanges: []httpRange{{0, 1}}},
	{r: "bytes=0-10", size: 1, expRanges: []httpRange{{0, 1}}},
	{r: "bytes=0-", size: 1, expRanges: []httpRange{{0, 1}}},
	{r: "bytes=-5", size: 1, expRanges: []httpRange{{0, 1}}},

	{r: "", size: 11, expRanges: nil},
	{r: "bytes=0-4", size: 11, expRanges: []httpRange{{0, 5}}},
	{r: "bytes=2-", size: 11, expRanges: []httpRange{{2, 9}}},
	{r: "bytes=-5", size: 11, expRanges: []httpRange{{6, 5}}},
	{r: "bytes=3-7", size: 11, expRanges: []httpRange{{3, 5}}},

	{r: "bytes=0-0,-2", size: 11, expRanges: []httpRange{{0, 1}, {9, 2}}},
	{r: "bytes=0-1,5-8", size: 11, expRanges: []httpRange{{0, 2}, {5, 4}}},
	{r: "bytes=0-1,5-", size: 11, expRanges: []httpRange{{0, 2}, {5, 6}}},
	{r: "bytes=5-1000", size: 11, expRanges: []httpRange{{5, 6}}},
	{r: "bytes=0-,1-,2-,3-,4-", size: 9, expRanges: []httpRange{{0, 9}, {1, 8}, {2, 7}, {3, 6}, {4, 5}}},
	{r: "bytes=0-9", size: 11, expRanges: []httpRange{{0, 10}}},
	{r: "bytes=0-10", size: 11, expRanges: []httpRange{{0, 11}}},
	{r: "bytes=0-11", size: 11, expRanges: []httpRange{{0, 11}}},
	{r: "bytes=0-12", size: 11, expRanges: []httpRange{{0, 11}}},
	{r: "bytes=10-11", size: 11, expRanges: []httpRange{{10, 1}}},
	{r: "bytes=10-", size: 11, expRanges: []httpRange{{10, 1}}},
	{r: "bytes=11-", size: 11, expErr: true},
	{r: "bytes=11-12", size: 11, expErr: true},
	{r: "bytes=12-12", size: 11, expErr: true},
	{r: "bytes=11-100", size: 11, expErr: true},
	{r: "bytes=12-100", size: 11, expErr: true},
	{r: "bytes=100-", size: 11, expErr: true},
	{r: "bytes=100-1000", size: 11, expErr: true},
}

func TestRequestRangeParsing(t *testing.T) {
	t.Parallel()
	for _, test := range reqRangeTests {
		res, err := parseReqRange(test.r, test.size)

		if err != nil && !test.expErr {
			t.Errorf("Received an unexpected error for test %s: %s", test, err)
		}
		if err == nil && test.expErr {
			t.Errorf("Expected to receive an error for test %s", test)
		}
		if !reflect.DeepEqual(res, test.expRanges) {
			t.Errorf("The received ranges for test %s '%#v' differ from the expected '%#v'", test, res, test.expRanges)
		}
	}
}
