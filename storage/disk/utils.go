package disk

import (
	"fmt"
	"path"

	"github.com/ironsmile/nedomi/types"
)

const (
	objectMetadataFileName = "objID"
)

func (s *Disk) getObjectIDPath(id *types.ObjectID) string {
	h := id.Hash()
	// Disk objects are writen 2 levels deep with maximum of 256 folders in each
	return fmt.Sprintf("%s/%x/%x/%x", s.path, h[0], h[1], h)
}

func (s *Disk) getObjectIndexPath(index *types.ObjectIndex) string {
	// For easier soring by hand, object parts are padded to 6 digits
	return fmt.Sprintf("%s/%06d", s.getObjectIDPath(index.ObjID), index.Part)
}

func (s *Disk) getObjectMetadataPath(omd *types.ObjectMetadata) string {
	return path.Join(s.getObjectIDPath(omd.ID), objectMetadataFileName)
}
