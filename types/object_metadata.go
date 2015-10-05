package types

import (
	"net/http"
)

// ObjectMetadata represents all the needed metadata of a cacheable object.
type ObjectMetadata struct {

	// The ObjectID for this file. It is used for identifying the object throughout
	// the cache layers in nedomi.
	ID *ObjectID

	// The time of the first request/response for this object as unix timestamp.
	ResponseTimestamp int64

	// Status code of the first proxied response for this object.
	Code int

	// The object size in bytes. Normally this should correspond to the
	// upstream's Content-Length header.
	Size uint64

	// HTTP headers which were received from the upstream and which we should
	// pass down for this object for any subsequent request.
	Headers http.Header

	// The time at wich this object can be considered stale. After this time
	// the object must be revalidated or discarded. This value is a unix timestamp.
	ExpiresAt int64
}
