package cache

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

const (
	input      = "The input of the test"
	objectPath = "testPathA"
	inputSize  = uint64(len(input))
)

func testCacheZone(id string, partSize uint64) types.CacheZone {
	return types.CacheZone{
		ID:      id,
		Storage: storage.NewMock(partSize),
		Algorithm: cache.NewMock(&cache.MockRepliers{
			ShouldKeep: func(*types.ObjectIndex) bool { return true },
		}),
	}
}

func TestPartWriter(t *testing.T) {
	t.Parallel()
	var cacheZoneID = "1"
	oid := types.NewObjectID(cacheZoneID, objectPath)
	oMeta := &types.ObjectMetadata{
		ID:                oid,
		ResponseTimestamp: time.Now().Unix(),
		Code:              200,
		Size:              inputSize,
		Headers:           nil,
	}

	test := func(start, length, partSize uint64) {
		cz := testCacheZone(cacheZoneID, partSize)
		cz.Storage.SaveMetadata(oMeta)
		pw := PartWriter(cz, oid,
			utils.HTTPContentRange{
				Start:   start,
				Length:  length,
				ObjSize: inputSize,
			})
		n, err := pw.Write([]byte(input[start : start+length]))
		if err != nil {
			t.Errorf("[PartWriter(%d,%d),Storage(%d)] returned error %s", start, length, partSize, err)
		}
		if uint64(n) != length {
			t.Errorf("[PartWriter(%d,%d),Storage(%d)] wrote %d expected %d", start, length, partSize, n, length)
		}

		if err := pw.Close(); err != nil {
			t.Error(err)
		}
		checkParts(t, start, length, cz.Storage, oid)
	}
	// Test all variants
	for partSize := uint64(1); partSize <= inputSize+3; partSize++ {
		for start := uint64(0); start < inputSize; start++ {
			for length := uint64(1); length+start <= inputSize; length++ {
				test(start, length, partSize)
			}
		}
	}
}

func checkParts(t *testing.T, start, length uint64, st types.Storage, oid *types.ObjectID) {
	partSize := st.PartSize()
	indexes := utils.BreakInIndexes(oid, start, start+length-1, partSize)
	prefix := fmt.Sprintf("[PartWriter(%d,%d),Storage(%d)]", start, length, partSize)

	if len(indexes) > 0 && start%partSize != 0 {
		checkPartIsMissing(t, prefix+"first part", st, indexes[0])
		indexes = indexes[1:]
	}
	if len(indexes) > 0 && (start+length)%partSize != 0 && (start+length) != inputSize {
		checkPartIsMissing(t, prefix+"last part", st, indexes[len(indexes)-1])
		indexes = indexes[:len(indexes)-1]
	}
	for _, idx := range indexes {
		partNum := uint64(idx.Part)
		expected := input[partSize*partNum : umin(inputSize, (partNum+1)*partSize)]
		checkPart(t, fmt.Sprintf("%s part %d", prefix, partNum), expected, st, idx)
	}
}

func checkPartIsMissing(t *testing.T, errPrefix string, st types.Storage, idx *types.ObjectIndex) {
	part, err := st.GetPart(idx)
	if os.IsNotExist(err) {
		return // all is fine
	} else if err != nil {
		t.Errorf("%s - was expected to be os.ErrNotExist but it was %s", errPrefix, err)
		return
	}
	t.Errorf("%s - was expected to be erronous but it wasn't", errPrefix)
	got, err := ioutil.ReadAll(part)
	if err != nil {
		t.Errorf("%s - %s", errPrefix, err)
		return
	}
	t.Errorf("%s has contents of `%s`", errPrefix, got)
}

func checkPart(t *testing.T, errPrefix string, expected string, st types.Storage, idx *types.ObjectIndex) {
	part, err := st.GetPart(idx)
	if err != nil {
		t.Errorf("%s - %s", errPrefix, err)
		return
	}
	got, err := ioutil.ReadAll(part)
	if err != nil {
		t.Errorf("%s - %s", errPrefix, err)
		return
	}
	if string(got) != expected {
		t.Errorf("%s expected \n`%s`\n, got \n`%s`", errPrefix, expected, string(got))
	}
}
