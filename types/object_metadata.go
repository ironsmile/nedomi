package types

import "net/http"

// ObjectMetadata represents all the needed metadata of a cachable object.
type ObjectMetadata struct {
	ID                *ObjectID
	ResponseTimestamp int64
	Code              int
	Size              uint64
	Headers           http.Header
}
