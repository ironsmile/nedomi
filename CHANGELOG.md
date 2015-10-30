# Change Log

A human readable change log between our released versions can be found in here.

## v0.1.1 - 2015-10-29

### New Stuff

 * More useful error messages on wrong syntax in the config file. An error will now be pointed out and some context will be printed.
 * All access log lines now contain identification for which vhost they refer to.
 * The purge module uses HTTP return code to signal that a purge was now unsuccessful for some reason.

### Bugfixes

 * The combination between "Date" and "Cache-Control: max-age" response headers is more RFC compliant. The Date will be stored as-is from the origin server and max-age will be calculated as time after this Date.
 * A bug in resizing down a cache zone with reload (`kill -HUP`) made it possible for this zone to be larger than expected. In some situations it resulted in crashes.

### Development

 * Starting with this one all release branches will have their dependencies vendored. For this end the `GO15VENDOREXPERIMENT` is used and it can be seen in the Makefile.
 * Support for `go1.4` is dropped.

## v0.1.0 - 2015-10-26

 * Initial Release
