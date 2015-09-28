package utils

import "github.com/ironsmile/nedomi/types"

//!TODO: find a better place than utils for this

// BreakInIndexes returns a slice of ObjectIndexes according to the specified
// byte range [start, end] (inclusive).
func BreakInIndexes(id *types.ObjectID, start, end, partSize uint64) []*types.ObjectIndex {
	firstIndex := start / partSize
	lastIndex := end/partSize + 1
	result := make([]*types.ObjectIndex, 0, lastIndex-firstIndex)
	for i := firstIndex; i < lastIndex; i++ {
		result = append(result, &types.ObjectIndex{
			ObjID: id,
			Part:  uint32(i),
		})
	}
	return result
}
