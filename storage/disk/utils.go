package disk

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ironsmile/nedomi/config"
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

func appendRandomSuffix(path string) string {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		panic(fmt.Sprintf("Could not read random data: %s", err))
	}

	return path + "_" + hex.EncodeToString(randBytes)
}

func (s *Disk) getObjectIDPath(id *types.ObjectID) string {
	// !TODO redo this with more []byte appending(we know how big it will be)
	// less string concatination
	h := id.StrHash()
	// Disk objects are written 2 levels deep with maximum of 256 folders in each
	if s.skipCacheKeyInPath {
		return filepath.Join(s.path, h[0:2], h[2:4], h)
	}

	return filepath.Join(s.path, id.CacheKey(), h[0:2], h[2:4], h)
}

func (s *Disk) getObjectIndexPath(idx *types.ObjectIndex) string {
	return filepath.Join(s.getObjectIDPath(idx.ObjID), getPartFilename(idx.Part))
}

func (s *Disk) getObjectMetadataPath(id *types.ObjectID) string {
	return filepath.Join(s.getObjectIDPath(id), objectMetadataFileName)
}

func (s *Disk) createFile(filePath string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), s.dirPermissions); err != nil {
		return nil, err
	}

	return os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, s.filePermissions)
}

func (s *Disk) getPartSize(partNum uint32, objectSize uint64) uint64 {
	//!TODO: move this in utils? delete?
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

func (s *Disk) checkPreviousDiskSettings(newSettings *config.CacheZone) error {
	f, err := os.Open(filepath.Join(s.path, diskSettingsFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	oldSettings := &config.CacheZone{}
	if err := json.NewDecoder(f).Decode(&oldSettings); err != nil {
		return utils.NewCompositeError(err, f.Close())
	}
	if err := f.Close(); err != nil {
		return err
	}

	if oldSettings.PartSize != newSettings.PartSize {
		return fmt.Errorf("Old partsize is %d and new partsize is %d",
			oldSettings.PartSize, newSettings.PartSize)
	}
	//!TODO: more validation?
	return nil
}

func (s *Disk) saveSettingsOnDisk(cz *config.CacheZone) error {
	if err := s.checkPreviousDiskSettings(cz); err != nil {
		return err
	}

	filePath := filepath.Join(s.path, diskSettingsFileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, s.filePermissions)
	if err != nil {
		return err
	}

	if err = json.NewEncoder(f).Encode(cz); err != nil {
		return utils.NewCompositeError(err, f.Close())
	}

	return f.Close()
}
