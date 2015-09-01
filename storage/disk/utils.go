package disk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

const (
	objectMetadataFileName = "objID"
	diskSettingsFileName   = ".nedomi-cache-storage"
)

func getPartFilename(part uint32) string {
	// For easier soring by hand, object parts are padded to 6 digits
	return fmt.Sprintf("%06d", part)
}

func (s *Disk) getObjectIDPath(id *types.ObjectID) string {
	h := id.Hash()
	// Disk objects are writen 2 levels deep with maximum of 256 folders in each
	return fmt.Sprintf("%s/%s/%x/%x/%x", s.path, id.CacheKey, h[0], h[1], h)
}

func (s *Disk) getObjectIndexPath(idx *types.ObjectIndex) string {
	return path.Join(s.getObjectIDPath(idx.ObjID), getPartFilename(idx.Part))
}

func (s *Disk) getObjectMetadataPath(id *types.ObjectID) string {
	return path.Join(s.getObjectIDPath(id), objectMetadataFileName)
}

func (s *Disk) createFile(filePath string) (*os.File, error) {
	if err := os.MkdirAll(path.Dir(filePath), s.dirPermissions); err != nil {
		return nil, err
	}

	return os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, s.filePermissions)
}

func (s *Disk) getPartSize(partNum uint32, objectSize uint64) uint64 {
	//!TODO: move this in utils?
	wholeParts := uint32(objectSize / s.partSize)
	remainder := objectSize % s.partSize
	if partNum > wholeParts {
		// Parts numbers start from 0, so partNum cannot be more than
		// wholeParts, even if there is a last smaller piece
		return 0
	} else if partNum == wholeParts {
		// Either there is a last smaller piece with size remainder or the file
		// was evenly split (remainder == 0) and there is no such piece
		return remainder
	}
	return s.partSize
}

func (s *Disk) getPartNumberFromFile(name string) (uint32, error) {
	partNum, err := strconv.Atoi(name)
	if err != nil || getPartFilename(uint32(partNum)) != name {
		return 0, fmt.Errorf("Invalid part filename '%s'", name)
	}

	return uint32(partNum), nil
}

func (s *Disk) getObjectMetadata(objPath string) (*types.ObjectMetadata, error) {
	f, err := os.Open(objPath)
	if err != nil {
		return nil, err
	}

	obj := &types.ObjectMetadata{}
	if err := json.NewDecoder(f).Decode(&obj); err != nil {
		return nil, utils.NewCompositeError(err, f.Close())
	}

	if filepath.Base(filepath.Dir(objPath)) != obj.ID.StrHash() {
		err := fmt.Errorf("The object %s was in the wrong directory: %s", obj.ID, objPath)
		return nil, utils.NewCompositeError(err, f.Close())
	}
	//!TODO: add more validation? ex. compare the cache key as well? also the
	// data itself may be corrupted or from an old app version

	return obj, f.Close()
}

func (s *Disk) getAvailableParts(obj *types.ObjectMetadata) (types.ObjectIndexMap, error) {
	files, err := ioutil.ReadDir(s.getObjectIDPath(obj.ID))
	if err != nil {
		return nil, err
	}

	parts := make(types.ObjectIndexMap)
	for _, f := range files {
		if f.Name() == objectMetadataFileName {
			continue
		}

		partNum, err := s.getPartNumberFromFile(f.Name())
		if err != nil {
			return nil, fmt.Errorf("Wrong part file for %s: %s", obj.ID, err)
		} else if s.getPartSize(partNum, obj.Size) != uint64(f.Size()) {
			return nil, fmt.Errorf("Part file %d for %s has incorrect size", partNum, obj.ID)
		} else {
			parts[partNum] = struct{}{}
		}
	}

	return parts, nil
}

func (s *Disk) checkPreviousDiskSettings() error {
	f, err := os.Open(path.Join(s.path, diskSettingsFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	oldSettings := &Disk{}
	if err := json.NewDecoder(f).Decode(&oldSettings); err != nil {
		return utils.NewCompositeError(err, f.Close())
	}
	if err := f.Close(); err != nil {
		return err
	}

	if oldSettings.partSize != s.partSize {
		return fmt.Errorf("Old partsize is %d and new partsize is %d", oldSettings.partSize, s.partSize)
	}
	return nil
}

func (s *Disk) saveSettingsOnDisk() error {
	if err := s.checkPreviousDiskSettings(); err != nil {
		return err
	}

	f, err := s.createFile(path.Join(s.path, diskSettingsFileName))
	if err != nil {
		return err
	}

	if err = json.NewEncoder(f).Encode(s); err != nil {
		return utils.NewCompositeError(err, f.Close())
	}

	return f.Close()
}