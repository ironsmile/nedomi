package cacheutils

import (
	"net/http"
	"strings"
	"time"

	"github.com/ironsmile/nedomi/types"
	"github.com/pquerna/cachecontrol/cacheobject"
)

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
func IsResponseCacheable(code int, headers http.Header) bool {
	//!TODO: write a better custom implementation or fork the cacheobject - the API sucks
	//!TODO: correctly handle cache-control, pragma, etag and vary headers
	//!TODO: write unit tests

	if code != http.StatusOK && code != http.StatusPartialContent {
		return false
	}

	// For now, we do not cache encoded responses
	if headers.Get("Content-Encoding") != "" {
		return false
	}

	// We do not cache multipart range responses
	if strings.Contains(headers.Get("Content-Type"), "multipart/byteranges") {
		return false
	}

	respDir, err := cacheobject.ParseResponseCacheControl(headers.Get("Cache-Control"))
	if err != nil || respDir.NoCachePresent || respDir.NoStore || respDir.PrivatePresent {
		return false
	}

	return true
}

// ResponseExpiresIn parses the expiration time from upstream headers, if any, and returns
// it as a duration from now. If no expire time is found, it returns its second argument:
// the default expiration time.
func ResponseExpiresIn(headers http.Header, ifNotAny time.Duration) time.Duration {

	//!TODO: this cacheobject.ParseResponseCacheControl is called two times for every
	// caceable response. We should find a way not to duplicate work. Maybe pass the resout
	// around somehow?
	respDir, err := cacheobject.ParseResponseCacheControl(headers.Get("Cache-Control"))
	if err != nil {
		return ifNotAny
	}

	if respDir.SMaxAge > 0 {
		return time.Duration(respDir.SMaxAge) * time.Second
	} else if respDir.MaxAge > 0 {
		return time.Duration(respDir.MaxAge) * time.Second
	} else if headers.Get("Expires") != "" {
		_ = "breakpoint"
		if t, err := time.Parse(time.RFC1123, headers.Get("Expires")); err == nil {
			//!TODO: use the server time from the Date header to calculate?
			return t.Sub(time.Now())
		}
	}

	return ifNotAny
}
