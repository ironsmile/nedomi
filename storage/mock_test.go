package storage

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/types"
)

var obj1 = &types.ObjectMetadata{
	ID:                types.NewObjectID("testkey", "/lorem/ipsum"),
	ResponseTimestamp: time.Now().Unix(),
	Size:              121,
	Headers:           http.Header{"test": []string{"mest"}},
}

var obj2 = &types.ObjectMetadata{
	ID:                types.NewObjectID("testkey", "/lorem/ipsum/2"),
	ResponseTimestamp: time.Now().Unix(),
	Size:              22,
	Headers:           http.Header{},
}

func TestMockStorageExpectedErrors(t *testing.T) {
	t.Parallel()
	s := NewMock()

	idx := &types.ObjectIndex{ObjID: obj1.ID, Part: 5}
	if _, err := s.GetMetadata(obj1.ID); !os.IsNotExist(err) {
		t.Errorf("Exptected to get os.ErrNotExist but got %#v", err)
	}
	if _, err := s.GetPart(idx); !os.IsNotExist(err) {
		t.Errorf("Exptected to get os.ErrNotExist but got %#v", err)
	}
	if err := s.SavePart(idx, strings.NewReader("0123456789")); err == nil {
		t.Error("Saving an index when no metadata is present should fail")
	}
}

func saveMetadata(t *testing.T, s *MockStorage, obj *types.ObjectMetadata) {
	if err := s.SaveMetadata(obj); err != nil {
		t.Fatalf("Could not save metadata for %s: %s", obj.ID, err)
	}
	if err := s.SaveMetadata(obj); !os.IsExist(err) {
		t.Errorf("Expected to get os.ErrExist but got %#v", err)
	}

	if res, ok := s.Objects[obj.ID.Hash()]; !ok || res != obj {
		t.Errorf("We should saved the same pointer: %p != %p", res, obj)
	}

	if res, err := s.GetMetadata(obj.ID); err != nil || res == nil {
		t.Errorf("Received unexpected error while getting metadata: %s", err)
	} else if res != obj {
		t.Errorf("We should have received the same pointer: %p != %p", res, obj)
	}
}

func savePart(t *testing.T, s *MockStorage, idx *types.ObjectIndex, contents string) {
	if err := s.SavePart(idx, strings.NewReader(contents)); err != nil {
		t.Fatalf("Could not save file part %s: %s", idx, err)
	}
	if err := s.SavePart(idx, strings.NewReader(contents)); !os.IsExist(err) {
		t.Errorf("Expected to get os.ErrExist but got %#v", err)
	}

	if bucket, ok := s.Parts[idx.ObjID.Hash()]; !ok {
		t.Errorf("Could not find the part bucket")
	} else if res, ok := bucket[idx.Part]; !ok || string(res) != contents {
		t.Errorf("Did not receive the same contents: %s", res)
	}

	if partReader, err := s.GetPart(idx); err != nil {
		t.Errorf("Received unexpected error while getting part: %s", err)
	} else if readContents, err := ioutil.ReadAll(partReader); err != nil {
		t.Errorf("Could not read saved part: %s", err)
	} else if string(readContents) != contents {
		t.Errorf("Expected the contents to be %s but read %s", contents, readContents)
	}
}

func TestMockStorageOperations(t *testing.T) {
	t.Parallel()
	s := NewMock()

	saveMetadata(t, s, obj1)
	saveMetadata(t, s, obj2)

	idx := &types.ObjectIndex{ObjID: obj2.ID, Part: 13}
	savePart(t, s, idx, "loremipsum2")

	passed := false
	s.Iterate(func(obj *types.ObjectMetadata, parts types.ObjectIndexMap) bool {
		if passed {
			t.Fatal("Expected iteration to stop after the first result")
		}
		passed = true
		return false
	})
	s.Discard(obj1.ID)
	if len(s.Objects) != 1 {
		t.Errorf("Expected only 1 remaining object but there are %d", len(s.Objects))
	}

	s.Iterate(func(obj *types.ObjectMetadata, parts types.ObjectIndexMap) bool {
		if obj != obj2 {
			t.Error("Expected to receive obj2's pointer")
		}
		if _, ok := parts[idx.Part]; !ok {
			t.Errorf("Expected part %s to be present", idx)
		}

		return false
	})

	s.DiscardPart(idx)
	s.Discard(obj2.ID)
	s.Iterate(func(obj *types.ObjectMetadata, parts types.ObjectIndexMap) bool {
		t.Error("Expected never to be called")
		return false
	})
}
