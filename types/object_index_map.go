package types

// ObjectIndexMap has a map that specifies which parts of a file are present.
type ObjectIndexMap struct {
	ObjID ObjectID
	Size  uint64
	Parts map[uint32]struct{}
}
