package utils

import (
	"testing"

	"github.com/ironsmile/nedomi/types"
)

type testCase struct {
	start    uint64
	end      uint64
	partSize uint64
	result   []uint32
}

var breakInIndexesMatrix = []testCase{
	{start: 0, end: 49, partSize: 50, result: []uint32{0}},
	{start: 0, end: 50, partSize: 50, result: []uint32{0, 1}},
	{start: 0, end: 99, partSize: 50, result: []uint32{0, 1}},
	{start: 5, end: 99, partSize: 50, result: []uint32{0, 1}},
	{start: 5, end: 100, partSize: 50, result: []uint32{0, 1, 2}},
	{start: 50, end: 99, partSize: 50, result: []uint32{1}},
	{start: 50, end: 50, partSize: 50, result: []uint32{1}},
	{start: 50, end: 49, partSize: 50, result: []uint32{}},
	{start: 0, end: 3, partSize: 1, result: []uint32{0, 1, 2, 3}},
}

func TestBreakInIndexes(t *testing.T) {
	t.Parallel()
	id := types.NewObjectID("test", "mest")
	for index, test := range breakInIndexesMatrix {
		var result = BreakInIndexes(id, test.start, test.end, test.partSize)
		if len(result) != len(test.result) {
			t.Errorf("Wrong len (%d != %d) on test index %d", len(result), len(test.result), index)
		}

		for resultIndex, value := range result {
			if value.Part != test.result[resultIndex] {
				t.Errorf("Wrong part for test index %d, wanted %d in position %d but got %d", index, test.result[resultIndex], resultIndex, value.Part)
			}
		}
	}
}
