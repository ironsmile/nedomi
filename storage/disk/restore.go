package disk

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

const metaFileName = "meta_file"

//!TODO this shoulde be redone when a more certain naming is imposed on files and directories
func (d *Disk) loadFromDisk() error {
	if meta, err := d.readMetaFromDisk(); err != nil || meta.PartSize != d.partSize {
		e := new(utils.CompositeError)
		e.AppendError(err)
		d.logger.Logf("%s is going to be removed.", d.path)
		e.AppendError(os.RemoveAll(d.path))
		return e
	}

	d.logger.Logf("%s is going to be loaded for reusage by storage.disk", d.path)
	walker := newDiskWalker(d)
	walker.walk()

	return nil
}

func (dw *diskWalker) loadWalk(path string, info os.FileInfo, err error) error {
	if info.IsDir() { // directories are not interesting
		return nil
	}
	if filepath.Base(path) == headerFileName { // headers are exempt for now
		return nil
	}

	rel, err := filepath.Rel(dw.disk.path, path) // this could be done faster
	if err != nil {                              // should never be possible
		return err
	}
	index, err := indexFromPath(rel)
	if err != nil {
		dw.disk.logger.Debugf("Got error from IndexFromPath - %s", err)
		return nil
	}

	expectedSize, err := dw.expectedSizeFor(index)
	if err != nil || int(info.Size()) != expectedSize || !dw.disk.cache.ShouldKeep(index) {
		if err != nil {
			dw.disk.logger.Errorf("Got error %s while getting expected size for %s", err, index)
		}

		dw.disk.DiscardIndex(index)
	}

	return nil
}

// mostly a place for sizes
type diskWalker struct {
	sizes map[types.ObjectID]int
	disk  *Disk
}

func (dw *diskWalker) expectedSizeFor(index types.ObjectIndex) (int, error) {
	fullSize, err := dw.getFullSizeFor(index.ObjID)
	if err != nil {
		return -1, err
	}
	if fullSize/int(dw.disk.partSize) == int(index.Part) {
		return fullSize % int(dw.disk.partSize), nil
	} else {
		return int(dw.disk.partSize), nil
	}

}

func (dw *diskWalker) getFullSizeFor(id types.ObjectID) (int, error) {
	if fullSize, ok := dw.sizes[id]; ok {
		return fullSize, nil
	}
	header, err := dw.disk.readHeaderFromFile(id)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(strings.TrimSpace(header.Get("Content-Length")))
}

func newDiskWalker(d *Disk) *diskWalker {
	return &diskWalker{
		sizes: make(map[types.ObjectID]int),
		disk:  d,
	}

}

func (dw *diskWalker) walk() error {
	return filepath.Walk(dw.disk.path, dw.loadWalk)
}

func indexFromPath(path string) (types.ObjectIndex, error) {
	dir, file := filepath.Split(path)
	part, err := strconv.Atoi(file)
	if err != nil {
		return types.ObjectIndex{}, fmt.Errorf("%s is not file for an ObjectIndex", path)
	}
	ID, err := iDFromPath(dir)
	if err != nil {
		return types.ObjectIndex{}, err
	}

	return types.ObjectIndex{
		ObjID: ID,
		Part:  uint32(part),
	}, nil
}

func iDFromPath(path string) (types.ObjectID, error) {
	for index, char := range path {
		if char == filepath.Separator {
			return types.ObjectID{
				CacheKey: path[0:index],
				Path:     path[index : len(path)-1],
				//!TODO: this will always return path without a final '/'
				// should probably be read from the headers
			}, nil

		}
	}
	return types.ObjectID{}, fmt.Errorf("%s is not path for an ID", path)
}

type meta struct {
	StorageObjects uint64 // not used
	PartSize       uint64
}

func (d *Disk) saveMetaToDisk() error {
	m := &meta{
		StorageObjects: d.storageObjects,
		PartSize:       d.partSize,
	}

	meta_file_path := path.Join(d.path, metaFileName)
	err := os.MkdirAll(path.Dir(meta_file_path), 0700)
	if err != nil {
		return fmt.Errorf("Error on creating the directory %s - %s", path.Dir(meta_file_path), err)
	}
	meta_file, err := os.Create(meta_file_path)
	if err != nil {
		return fmt.Errorf("Error on creating meta file for disk storage - %s", err)
	}
	defer meta_file.Close()

	err = json.NewEncoder(meta_file).Encode(m)

	if err != nil {
		return fmt.Errorf("Error on encoding to meta file - %s", err)
	}

	d.logger.Logf("Wrote meta to %s", path.Join(d.path, metaFileName))
	return nil
}

func (d *Disk) readMetaFromDisk() (*meta, error) {
	m := new(meta)
	meta_file, err := os.Open(path.Join(d.path, metaFileName))
	if err != nil {
		return nil, fmt.Errorf("Error on opening meta file for disk storage - %s", err)
	}
	defer meta_file.Close()

	err = json.NewDecoder(meta_file).Decode(m)

	if err != nil {
		return nil, fmt.Errorf("Error on decoding from meta file - %s", err)
	}

	return m, nil
}
