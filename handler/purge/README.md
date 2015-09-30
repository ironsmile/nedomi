#Purge

##Configuration:
no configuration is required for the handler

##API:

Make a POST request with the following body to *any* URL handled by the purge handler, with a body as follows:

```json
{
	"cache_zone":"a zone name",
	"cache_zone_key": "a key for a zone",
	"objects" : [
		"/path/to/a/file/to/be/purged",
		"/path/to/another/file"
	]
}
```

the request shown will remove a file with the first name and all files that start with `/path/to/a/files`.

The returned result will be of the form:

```json
{
	"cache_zone":"a zone name",
	"cache_zone_key": "a key for a zone",
	"results" : {
		"/path/to/a/file/to/be/purged": true,
		"/path/to/another/file": false
	}
}
```

the map in the result will have for value true if files have been deleted and false otherwise.

##TODO:

* async api with meaningful urls
* glob matching
* authentication of any kind
