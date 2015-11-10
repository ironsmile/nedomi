# Change Log

A human readable change log between our released versions can be found in here.

## v0.1.5 - 2015-11-10

### New Stuff

* Nedomi can now utilise more efficient OS syscalls for copying data in some situations. For example, with cached files in Linux the `sendfile` syscall will be used to bridge the opened file and network socket directly in kernel space. For other systems, their respectful method for efficient copying will be used. This behaviour is made possible by the golang's standard library.

## v0.1.4 - 2015-11-09

### New Stuff

* Updated the [mp4 lib](https://github.com/MStoykov/mp4) used for pseudo streaming to its version 0.1.3. The previous version had a bug which was leading to premanently blocked go routines.

## v0.1.3 - 2015-11-05

### Bug fixes

* Read and write time outs (as set by `read_timeout` and `write_timeout`) were actually deadlines. Requests were expected to finish *all* of their reading and writing in the specified number of seconds. This is now changed to actual time outs on the connection. A time out will happen only if the connection is actually stalled. Reading or writing a single byte every `read/write_timeout` period means the connection will not be marked as stalled.

* Pulls a bug fix from the upstream mp4 library. In some situations the atom `udta' was not parsed correctly. The effect of this prematurely terminated request. 

## v0.1.2 - 2015-11-03

### New Stuff

* The `skip_cache_key_in_path` option is added for a cache zone. By default it is `false` and the old behaviour for cache paths is used: inside the cache directory there will be one directory for every cache key used in this cache zone. When set to `true` no "cache key directory" will be used and all cached files will be rooted in the cache zone directory.

* The `ketama` upstream balancing algorithm is renamed to `legacyketama`. We are planning a better implementation of the algorithm but this one is still in use in some installations and must be kept for backward interoperability. **Important!** This change requires a change in the configuration file between v0.1.(0|1) and this one: `ketama` must become `legacyketama`.

* New option `upstream_hash_prefix` in the proxy handler. It is a string which will be used in front of the URIs as a key for the consistent hashing upstream balancing algorithms.

### Removed

* The `jump` upstream balancing algorithm is removed. It turns out it is not suitable for our purposes at all.

### Bug fixes

* Numerous bug fixes in the balancing algorithms in corner cases. They all use a thread-safe random number generator now.

### Development

* The severity of many error messages is lowered to info or debug.

## v0.1.1 - 2015-10-29

### New Stuff

 * More useful error messages on wrong syntax in the config file. An error will now be pointed out and some context will be printed.
 * All access log lines now contain identification for which vhost they refer to.
 * The purge module uses HTTP return code to signal that a purge was now unsuccessful for some reason.

### Bug fixes

 * The combination between "Date" and "Cache-Control: max-age" response headers is more RFC compliant. The Date will be stored as-is from the origin server and max-age will be calculated as time after this Date.
 * A bug in resizing down a cache zone with reload (`kill -HUP`) made it possible for this zone to be larger than expected. In some situations it resulted in crashes.

### Development

 * Starting with this one all release branches will have their dependencies vendored. For this end the `GO15VENDOREXPERIMENT` is used and it can be seen in the Makefile.
 * Support for `go1.4` is dropped.

## v0.1.0 - 2015-10-26

 * Initial Release
