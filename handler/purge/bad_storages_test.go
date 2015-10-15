package purge

import (
	"errors"
	"testing"

	"github.com/ironsmile/nedomi/types"
)

var errTest = errors.New("error for tests")

type badGetParts struct {
	types.Storage
	*testing.T
	badObjs []*types.ObjectID
}

func newBadGetParts(t *testing.T, storage types.Storage, objs ...*types.ObjectID) *badGetParts {
	return &badGetParts{
		T:       t,
		Storage: storage,
		badObjs: objs,
	}
}

func (bgp *badGetParts) GetAvailableParts(id *types.ObjectID) ([]*types.ObjectIndex, error) {
	for _, badObj := range bgp.badObjs {
		if *badObj == *id {
			return nil, errTest
		}
	}

	return bgp.Storage.GetAvailableParts(id)
}

type badDiscard struct {
	types.Storage
	*testing.T
	badObjs []*types.ObjectID
}

func newBadDiscard(t *testing.T, storage types.Storage, objs ...*types.ObjectID) *badDiscard {
	return &badDiscard{
		T:       t,
		Storage: storage,
		badObjs: objs,
	}
}

func (bd *badDiscard) Discard(id *types.ObjectID) error {
	for _, badObj := range bd.badObjs {
		if *badObj == *id {
			return errTest
		}
	}

	return bd.Storage.Discard(id)
}
