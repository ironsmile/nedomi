package cache

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/httputils"
)

const (
	input     = "The input of the test"
	inputSize = uint64(len(input))
)

func init() {
	rand.Seed(time.Now().Unix())
}

var oid = types.NewObjectID("testKey1", "testPathA")
var oMeta = &types.ObjectMetadata{
	ID:                oid,
	ResponseTimestamp: time.Now().Unix(),
	Code:              200,
	Size:              inputSize,
	Headers:           nil,
}

func TestPartWriter(t *testing.T) {
	t.Parallel()
	test := func(start, length, partSize uint64) {
		cz := &types.CacheZone{
			ID:      "TestCZ",
			Storage: storage.NewMock(partSize),
			Algorithm: cache.NewMock(&cache.MockRepliers{
				ShouldKeep: func(*types.ObjectIndex) bool { return true },
			}),
		}

		if err := cz.Storage.SaveMetadata(oMeta); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		write(t, start, length, cz, oid)
		checkParts(t, start, length, cz, oid)
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

func TestAlgorithmCompliance(t *testing.T) {
	t.Parallel()
	partSize := uint64(5)
	totalParts := inputSize/partSize + 1

	state := make([]struct{ added, promoted bool }, totalParts)
	expectedParts := make([]bool, totalParts)

	algo := cache.NewMock(&cache.MockRepliers{
		AddObject: func(idx *types.ObjectIndex) error {
			state[idx.Part].added = true
			return nil
		},
		PromoteObject: func(idx *types.ObjectIndex) {
			state[idx.Part].promoted = true
		},
	})
	for i := uint64(0); i < totalParts; i++ {
		if rand.Intn(2) == 1 {
			expectedParts[i] = true
			algo.SetFakeReplies(&types.ObjectIndex{ObjID: oid, Part: uint32(i)}, &cache.MockRepliers{
				ShouldKeep: func(*types.ObjectIndex) bool { return true },
			})
		}
	}
	cz := &types.CacheZone{ID: "TestCZ", Storage: storage.NewMock(partSize), Algorithm: algo}
	if err := cz.Storage.SaveMetadata(oMeta); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	write(t, 0, inputSize, cz, oid)
	for i := uint64(0); i < totalParts; i++ {
		idx := &types.ObjectIndex{ObjID: oid, Part: uint32(i)}
		if expectedParts[i] {
			expected := input[partSize*i : umin(inputSize, (i+1)*partSize)]
			checkPart(t, fmt.Sprintf("part %d", i), expected, cz.Storage, idx)
			if !state[i].added || !state[i].promoted {
				t.Errorf("Wrong cache state for part %d: %v", i, state[i])
			}
		} else {
			checkPartIsMissing(t, fmt.Sprintf("part %d", i), cz.Storage, idx)
			if state[i].added || state[i].promoted {
				t.Errorf("Wrong cache state for expected missing part %d: %v", i, state[i])
			}
		}
	}
}

func write(t *testing.T, start, length uint64, cz *types.CacheZone, oid *types.ObjectID) {
	partSize := cz.Storage.PartSize()
	pw := PartWriter(cz, oid,
		httputils.ContentRange{
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

}

func checkParts(t *testing.T, start, length uint64, cz *types.CacheZone, oid *types.ObjectID) {
	partSize := cz.Storage.PartSize()
	indexes := utils.BreakInIndexes(oid, start, start+length-1, partSize)
	prefix := fmt.Sprintf("[PartWriter(%d,%d),Storage(%d)]", start, length, partSize)

	if len(indexes) > 0 && start%partSize != 0 {
		checkPartIsMissing(t, prefix+"first part", cz.Storage, indexes[0])
		indexes = indexes[1:]
	}
	if len(indexes) > 0 && (start+length)%partSize != 0 && (start+length) != inputSize {
		checkPartIsMissing(t, prefix+"last part", cz.Storage, indexes[len(indexes)-1])
		indexes = indexes[:len(indexes)-1]
	}
	for _, idx := range indexes {
		partNum := uint64(idx.Part)
		expected := input[partSize*partNum : umin(inputSize, (partNum+1)*partSize)]
		checkPart(t, fmt.Sprintf("%s part %d", prefix, partNum), expected, cz.Storage, idx)
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
