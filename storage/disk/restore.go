package disk

/*
// list of special names
var specialNames = []string{
	metaFileName,
	headerFileName,
	objectIDFileName,
}

const (
	metaFileName          = "meta_file"
	objectIDsPerIteration = 10
	maxStatsPerObjectID   = 10
)

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
	root, err := os.Open(d.path)
	if err != nil {
		e := new(utils.CompositeError)
		e.AppendError(err)
		d.logger.Logf("%s is going to be removed.", d.path)
		e.AppendError(os.RemoveAll(d.path))
		return err
	}

	infos, err := root.Readdir(-1)
	if err != nil {
		d.logger.Logf("Got error while reading the files in the storage root - %s", err)
	}

	for _, info := range infos {
		if info.IsDir() { // cache_key
			file, err := os.Open(filepath.Join(d.path, info.Name()))
			if err != nil {
				d.logger.Logf("Got error while loading cache for key %s:\n%s", info.Name(), err)
				continue
			}
			walker := newDiskWalker(d, file)
			walker.Walk()
		}
	}

	return nil
}

func (dw *diskWalker) Walk() {
	for {
		infos, err := dw.dir.Readdir(objectIDsPerIteration)
		if err != nil {
			if io.EOF != err {
				dw.disk.logger.Errorf("Error while reading directory %s to restore disk storage:\n %s", dw.dir.Name(), err)
			}
			return
		}

		for _, info := range infos {
			if info.IsDir() {
				if err := dw.restoreObjectID(info.Name()); err != nil {
					dw.disk.logger.Errorf("Got error while restoring ObjectID for hash %s:\n%s", info.Name(), err)
				}

			}
		}
	}
}

func (dw *diskWalker) restoreObjectID(objectIDhash string) error {
	id, err := dw.disk.readObjectIDForKeyNHash(path.Base(dw.dir.Name()), objectIDhash)
	if err != nil {
		//!TODO delete this dir?
		return err
	}

	objectIDdir, err := os.Open(filepath.Join(dw.disk.path, path.Base(dw.dir.Name()), objectIDhash))
	if err != nil {
		return err
	}
	defer objectIDdir.Close()

	for {
		infos, err := objectIDdir.Readdir(maxStatsPerObjectID)
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
		fullSize, err := dw.getFullSizeFor(id)
		if err != nil {
			//!TODO delete dir
			dw.disk.logger.Errorf("Error while finding out the fullSize of a objectID %s to be restored:\n%s", id, err)
			continue
		}

		for _, info := range infos {
			indexNumber, err := strconv.Atoi(info.Name())
			if err != nil {
				if !isSpecialFileName(info.Name()) {
					dw.disk.logger.Errorf("Parsing a supposive index file %s got error %s", info.Name(), err)
				}
				continue
			}
			index := types.ObjectIndex{
				ObjID: id,
				Part:  uint32(indexNumber),
			}
			expectedSize, err := dw.calculateExpectedSize(fullSize, indexNumber)
			if err != nil || int(info.Size()) != expectedSize || !dw.disk.cache.ShouldKeep(index) {
				if err != nil {
					dw.disk.logger.Errorf("Got error %s while getting expected size for %s", err, index)
				}

				dw.disk.DiscardIndex(index)
			}

		}
	}
}

func isSpecialFileName(name string) bool {
	for _, specialName := range specialNames {
		if name == specialName {
			return true
		}
	}
	return false
}

// mostly a place for sizes
type diskWalker struct {
	sizes map[types.ObjectID]int
	dir   *os.File
	disk  *Disk
}

func (dw *diskWalker) calculateExpectedSize(fullSize, index int) (int, error) {
	if fullSize/int(dw.disk.partSize) == index {
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

func newDiskWalker(d *Disk, dir *os.File) *diskWalker {
	return &diskWalker{
		sizes: make(map[types.ObjectID]int),
		disk:  d,
		dir:   dir,
	}

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
*/
