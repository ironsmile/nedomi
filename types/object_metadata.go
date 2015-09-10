package types

import "net/http"

// ObjectMetadata represents all the needed metadata of a cachable object.
type ObjectMetadata struct {
	ID                *ObjectID
	ResponseTimestamp int64
	Code              int
	Headers           http.Header
}
