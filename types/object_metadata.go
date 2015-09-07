package types

import (
	"net/http"
	"time"
)

// ObjectMetadata represents all the needed metadata of a cachable object.
type ObjectMetadata struct {
	ID           *ObjectID
	ResponseTime time.Time
	Size         uint64
	Code         int
	Headers      http.Header
}
