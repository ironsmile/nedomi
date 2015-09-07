package storage

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/ironsmile/nedomi/types"
)

//!TODO: Actually, looking at this, it is not a very good mock storage but it's
// an ok simple memory storage... maybe move this and get a real mock storage?

type key [types.ObjectIDHashSize]byte

// MockStorage implements the storage interface and is used for testing
type MockStorage struct {
	Objects map[key]*types.ObjectMetadata
	Parts   map[key]map[uint32][]byte
}

// GetMetadata returns the metadata for this object, if present.
func (s *MockStorage) GetMetadata(id *types.ObjectID) (*types.ObjectMetadata, error) {
	if obj, ok := s.Objects[id.Hash()]; ok {
		return obj, nil
	}
	return nil, os.ErrNotExist
}

// GetPart returns an io.ReadCloser that will read the specified part of the object.
func (s *MockStorage) GetPart(idx *types.ObjectIndex) (io.ReadCloser, error) {
	if obj, ok := s.Parts[idx.ObjID.Hash()]; ok {
		if part, ok := obj[idx.Part]; ok {
			return ioutil.NopCloser(bytes.NewReader(part)), nil
		}
	}
	return nil, os.ErrNotExist
}

// SaveMetadata saves the supplied metadata.
func (s *MockStorage) SaveMetadata(m *types.ObjectMetadata) error {
	if _, ok := s.Objects[m.ID.Hash()]; ok {
		return os.ErrExist
	}

	s.Objects[m.ID.Hash()] = m

	return nil
}

// SavePart saves the contents of the supplied object part.
func (s *MockStorage) SavePart(idx *types.ObjectIndex, data io.Reader) error {
	objHash := idx.ObjID.Hash()
	if _, ok := s.Objects[objHash]; !ok {
		return errors.New("Object metadata is not present")
	}

	if _, ok := s.Parts[objHash]; !ok {
		s.Parts[objHash] = make(map[uint32][]byte)
	}

	if _, ok := s.Parts[objHash][idx.Part]; ok {
		return os.ErrExist
	}

	contents, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	s.Parts[objHash][idx.Part] = contents
	return nil
}

// Discard removes the object and its metadata.
func (s *MockStorage) Discard(id *types.ObjectID) error {
	delete(s.Objects, id.Hash())
	delete(s.Parts, id.Hash())
	return nil
}

// DiscardPart removes the specified part of the object.
func (s *MockStorage) DiscardPart(idx *types.ObjectIndex) error {
	if obj, ok := s.Parts[idx.ObjID.Hash()]; ok {
		delete(obj, idx.Part)
	}
	return nil
}

// Iterate iterates over all the objects and passes them to the supplied callback
// function. If the callback function returns false, the iteration stops.
func (s *MockStorage) Iterate(callback func(*types.ObjectMetadata, types.ObjectIndexMap) bool) error {
	for hash, obj := range s.Objects {
		parts := make(types.ObjectIndexMap)
		for partNum := range s.Parts[hash] {
			parts[partNum] = struct{}{}
		}
		if !callback(obj, parts) {
			return nil
		}
	}
	return nil
}

// NewMock returns a new mock storage that ready for use.
func NewMock() *MockStorage {
	return &MockStorage{
		Objects: make(map[key]*types.ObjectMetadata),
		Parts:   make(map[key]map[uint32][]byte),
	}
}
