package disk

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"path"
	"strconv"

	"github.com/ironsmile/nedomi/types"
)

func pathFromIndex(index types.ObjectIndex) string {
	return path.Join(pathFromID(index.ObjID), strconv.Itoa(int(index.Part)))
}

func pathFromID(id types.ObjectID) string {
	h := md5.New()
	io.WriteString(h, id.Path)
	return path.Join(id.CacheKey, hex.EncodeToString(h.Sum(nil)))
}

func objectIDFileNameFromID(id types.ObjectID) string {
	return path.Join(pathFromID(id), objectIDFileName)
}

func headerFileNameFromID(id types.ObjectID) string {
	return path.Join(pathFromID(id), headerFileName)
}
