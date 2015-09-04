// Package utils exports few handy functions
package utils

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
)

// FileExists returns true if filePath is already existing regular file. If it is a
// directory FileExists will return false.
func FileExists(filePath string) bool {
	st, err := os.Stat(filePath)
	return err == nil && !st.IsDir()
}

// IsRequestCacheable returs whether the client allows the requested content to
// be retrieved from the cache. True result and unlimited duration means that the
func IsRequestCacheable(req *http.Request) bool {
	//!TODO: improve; implement something like github.com/pquerna/cachecontrol but better
	//!TODO: write unit tests

	reqDir, _ := cacheobject.ParseRequestCacheControl(req.Header.Get("Cache-Control"))
	return !(reqDir.NoCache || reqDir.NoStore)
}

// IsResponseCacheable returs whether the upstream server allows the requested
// content to be saved in the cache. True result and 0 duration means that the
// response has no expiry date.
func IsResponseCacheable(resp *http.Response) (bool, time.Duration) {
	//!TODO: write a better custom implementation or fork the cacheobject - the API sucks
	//!TODO: correctly handle cache-control, pragma, etag and vary headers
	//!TODO: write unit tests

	respDir, _ := cacheobject.ParseResponseCacheControl(resp.Header.Get("Cache-Control"))
	return !(respDir.NoCachePresent || respDir.NoStore || respDir.PrivatePresent), 0
}

// AddNewLineIfMissing returns the same string if there is '\n' at the end
// or a new string with '\n' of the input if the input does not end in '\n'
func AddNewLineIfMissing(input string) string {
	if strings.HasSuffix(input, `\n`) {
		return input
	}
	return input + `\n`
}
