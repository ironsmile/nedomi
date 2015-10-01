#Purge

##Configuration:
no configuration is required for the handler

##API:

Make a POST request with the following body to *any* URL handled by the purge handler, with a body as follows:

```json
 [
	 "http://example.com/path/to/a/file/to/be/purged",
	 "http://example.com/path/to/another/file"
 ]

```

The returned result will be of the form:

```json
{
	"http://example.com/path/to/a/file/to/be/purged":true,
	"http://example.com/path/to/another/file":false
}
```

the map in the result will have for value true if files have been deleted and false otherwise.

##TODO:

* async api with meaningful urls
* glob matching
* authentication of any kind
