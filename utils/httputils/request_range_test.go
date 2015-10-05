package httputils

import (
	"fmt"
	"reflect"
	"testing"
)

type reqRangeTest struct {
	r         string
	size      uint64
	expErr    bool
	expRanges []HTTPRange
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

	{r: "bytes=0-0", size: 1, expRanges: []HTTPRange{{0, 1}}},
	{r: "bytes=0-10", size: 1, expRanges: []HTTPRange{{0, 1}}},
	{r: "bytes=0-", size: 1, expRanges: []HTTPRange{{0, 1}}},
	{r: "bytes=-5", size: 1, expRanges: []HTTPRange{{0, 1}}},

	{r: "", size: 11, expRanges: nil},
	{r: "bytes=0-4", size: 11, expRanges: []HTTPRange{{0, 5}}},
	{r: "bytes=2-", size: 11, expRanges: []HTTPRange{{2, 9}}},
	{r: "bytes=-5", size: 11, expRanges: []HTTPRange{{6, 5}}},
	{r: "bytes=3-7", size: 11, expRanges: []HTTPRange{{3, 5}}},

	{r: "bytes=0-0,-2", size: 11, expRanges: []HTTPRange{{0, 1}, {9, 2}}},
	{r: "bytes=0-1,5-8", size: 11, expRanges: []HTTPRange{{0, 2}, {5, 4}}},
	{r: "bytes=0-1,5-", size: 11, expRanges: []HTTPRange{{0, 2}, {5, 6}}},
	{r: "bytes=5-1000", size: 11, expRanges: []HTTPRange{{5, 6}}},
	{r: "bytes=0-,1-,2-,3-,4-", size: 9, expRanges: []HTTPRange{{0, 9}, {1, 8}, {2, 7}, {3, 6}, {4, 5}}},
	{r: "bytes=0-9", size: 11, expRanges: []HTTPRange{{0, 10}}},
	{r: "bytes=0-10", size: 11, expRanges: []HTTPRange{{0, 11}}},
	{r: "bytes=0-11", size: 11, expRanges: []HTTPRange{{0, 11}}},
	{r: "bytes=0-12", size: 11, expRanges: []HTTPRange{{0, 11}}},
	{r: "bytes=10-11", size: 11, expRanges: []HTTPRange{{10, 1}}},
	{r: "bytes=10-", size: 11, expRanges: []HTTPRange{{10, 1}}},
	{r: "bytes=11-", size: 11, expErr: true},
	{r: "bytes=11-12", size: 11, expErr: true},
	{r: "bytes=12-12", size: 11, expErr: true},
	{r: "bytes=11-100", size: 11, expErr: true},
	{r: "bytes=12-100", size: 11, expErr: true},
	{r: "bytes=100-", size: 11, expErr: true},
	{r: "bytes=100-1000", size: 11, expErr: true},
}

func testReverse(t *testing.T, test reqRangeTest, ranges []HTTPRange) {
	for _, rng := range ranges {
		rev, err := ParseRespContentRange(rng.ContentRange(test.size))
		if err != nil {
			t.Errorf("Received an unexpected error for parsing the generated content range: %s", err)
		}
		if rev.ObjSize != test.size || rev.Start != rng.Start || rev.Length != rng.Length {
			t.Errorf("Mismatch between range %#v and generated content-range %#v for test %s", rng, rev, test)
		}
	}
}

func TestRequestRangeParsing(t *testing.T) {
	t.Parallel()
	for _, test := range reqRangeTests {
		ranges, err := ParseReqRange(test.r, test.size)

		if err != nil && !test.expErr {
			t.Errorf("Received an unexpected error for test %s: %s", test, err)
			continue
		}
		if err == nil && test.expErr {
			t.Errorf("Expected to receive an error for test %s", test)
			continue
		}
		if !reflect.DeepEqual(ranges, test.expRanges) {
			t.Errorf("The received ranges for test %s '%#v' differ from the expected '%#v'", test, ranges, test.expRanges)
		}
		testReverse(t, test, ranges)
	}
}
