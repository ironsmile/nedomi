package types

// ObjectFullMetadata contains everything from ObjectMetadata and also contains
// a map with the available parts of the object.
type ObjectFullMetadata struct {
	ObjectMetadata
	Parts ObjectIndexMap
}
