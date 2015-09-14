// Package utils exports few handy functions
package utils

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ironsmile/nedomi/types"
	"github.com/pquerna/cachecontrol/cacheobject"
)

// FileExists returns true if filePath is already existing regular file. If it is a
// directory FileExists will return false.
func FileExists(filePath string) bool {
	st, err := os.Stat(filePath)
	return err == nil && !st.IsDir()
}

// CacheSatisfiesRequest returs whether the client allows the requested content
// to be retrieved from the cache and whether the cache we have is fresh enough
// to be used for handling the request.
func CacheSatisfiesRequest(obj *types.ObjectMetadata, req *http.Request) bool {
	//!TODO: improve; implement something like github.com/pquerna/cachecontrol but better
	//!TODO: write unit tests

	//!TODO: handle cases with respect to https://tools.ietf.org/html/rfc7232

	reqDir, _ := cacheobject.ParseRequestCacheControl(req.Header.Get("Cache-Control"))
	return !(reqDir.NoCache || reqDir.NoStore)
}

// IsResponseCacheable returs whether the upstream server allows the requested
// content to be saved in the cache. True result and 0 duration means that the
// response has no expiry date.
func IsResponseCacheable(code int, headers http.Header) (bool, time.Duration) {
	//!TODO: write a better custom implementation or fork the cacheobject - the API sucks
	//!TODO: correctly handle cache-control, pragma, etag and vary headers
	//!TODO: write unit tests

	respDir, _ := cacheobject.ParseResponseCacheControl(headers.Get("Cache-Control"))

	if code != 200 && code != 206 {
		return false, 0
	}

	// For now, we do not cache encoded responses
	if headers.Get("Content-Encoding") != "" {
		return false, 0
	}

	// We do not cache multipart range responses
	if strings.Contains(headers.Get("Content-Type"), "multipart/byteranges") {
		return false, 0
	}

	return !(respDir.NoCachePresent || respDir.NoStore || respDir.PrivatePresent), 0
}

// IsMetadataFresh checks whether the supplied metadata could still be used.
func IsMetadataFresh(obj *types.ObjectMetadata) bool {
	//!TODO: implementation, tests
	return true
}
