package cache

import (
	"io"
	"net/http"

	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
)

func (c *CachingProxy) getPartReader(req *http.Request, objID *types.ObjectID) io.ReadCloser {
	//!TODO: handle ranges correctly, handle *unknown* object size correctly
	readers := []io.ReadCloser{}
	var part uint32
	for {
		idx := &types.ObjectIndex{ObjID: objID, Part: part}
		c.Logger.Debugf("[%p] Trying to load part %s from storage...", req, idx)
		r, err := c.Cache.Storage.GetPart(idx)
		if err != nil {
			break //TODO: fix, this is wrong
		}
		c.Logger.Debugf("[%p] Loaded part %s from storage!", req, idx)
		readers = append(readers, r)
		part++
	}
	c.Logger.Debugf("[%p] Return reader with %d parts of %s from storage!", req, len(readers), objID)
	return storage.MultiReadCloser(readers...)
}

func (c *CachingProxy) isMetadataFresh(obj *types.ObjectMetadata) bool {
	//!TODO: implement
	return true
}
