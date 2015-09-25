package cache

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

const (
	testPartSize = 5

	input      = "The input of the partwriter test. Some more text to not run out of test material"
	objectPath = "testPathA"
	inputSize  = len(input)
)

func testCacheZone(id string) types.CacheZone {
	st := storage.NewMock(testPartSize)
	al := cache.NewMock(&cache.MockReplies{
		ShouldKeep: true,
	})
	cz := types.CacheZone{
		ID:        id,
		Storage:   st,
		Algorithm: al,
	}
	return cz
}

func TestPartWriter(t *testing.T) {
	t.Parallel()
	var cacheZoneID = "1"
	oid := types.NewObjectID(cacheZoneID, objectPath)
	oMeta := &types.ObjectMetadata{
		ID:                oid,
		ResponseTimestamp: time.Now().Unix(),
		Code:              200,
		Size:              uint64(inputSize),
		Headers:           nil,
	}

	test := func(index int) {
		cz := testCacheZone(cacheZoneID)
		cz.Storage.SaveMetadata(oMeta)
		start := rand.Intn(inputSize / 2)
		length := rand.Intn(inputSize/4) + inputSize/4
		pw := PartWriter(cz, oid,
			utils.HTTPContentRange{
				Start:   uint64(start),
				Length:  uint64(length),
				ObjSize: uint64(inputSize),
			})
		n, err := pw.Write([]byte(input[start : start+length]))
		if err != nil {
			t.Errorf("%d: PartWriter(%d:%d) returned error %s", index, start, length, err)
		}
		if n != length {
			t.Errorf("%d: PartWriter(%d:%d) wrote %d expected %d", index, start, length, n, length)
		}

		if err := pw.Close(); err != nil {
			t.Error(err)
		}
		checkParts(t, start, length, cz.Storage, oid)
	}
	for index := 0; 1000 > index; index++ {
		go test(index)
	}
}

func checkParts(t *testing.T, start, length int, st types.Storage, oid *types.ObjectID) {
	lastPartNum := (start + length) / testPartSize
	firstPartNum := start / testPartSize
	if start%testPartSize != 0 {
		checkPartIsMissing(t, "first part", st, &types.ObjectIndex{
			ObjID: oid,
			Part:  uint32(firstPartNum),
		})
		firstPartNum++
	}
	if (start+length)%testPartSize != 0 && (start+length) != inputSize {
		checkPartIsMissing(t, "last part", st, &types.ObjectIndex{
			ObjID: oid,
			Part:  uint32(lastPartNum),
		})
		lastPartNum--
	}
	for i := firstPartNum; lastPartNum > i; i++ {
		expected := input[testPartSize*i : (i+1)*testPartSize]
		errPrefix := fmt.Sprintf("part %d:", i)
		checkPart(t, errPrefix, expected, st, &types.ObjectIndex{
			ObjID: oid,
			Part:  uint32(i),
		})
	}

}

func checkPartIsMissing(t *testing.T, errPrefix string, st types.Storage, objectIndex *types.ObjectIndex) {
	part, err := st.GetPart(objectIndex)
	if err != nil {
		return // all is fine
	}
	t.Errorf("%s - was expected to be erronous but it wasn't", errPrefix)
	got, err := ioutil.ReadAll(part)
	if err != nil {
		t.Errorf("%s - %s", errPrefix, err)
		return
	}
	t.Errorf("%s has contents of `%s`", errPrefix, got)
}

func checkPart(t *testing.T, errPrefix string, expected string, st types.Storage, objectIndex *types.ObjectIndex) {
	part, err := st.GetPart(objectIndex)
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
