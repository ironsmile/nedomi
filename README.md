# nedomi

HTTP media cache server. Most caching servers does not understand when a media file is being proxied. When storing media files you can gain an performance and space advantage when consider the fact that most of the big meda files are not actually watched from end to end.

We intend to implement a caching algorithm which takes all this into consideration and delivers better cache performance and throughput.

## Requirements

Nothing. It is pure Go. If you have some of [latest versions](https://golang.org/dl/) of the language you are good to go.

## Install

```
go get github.com/gophergala/nedomi
```

## Configuration

We all know that a mark of a good software is its configurability. Not everyone has the same needs. So we have you covered. nedomi lets you choose all the details.

The configuration is stored in a single [JSON file](https://en.wikipedia.org/wiki/JSON). It is basically an object with a key for every single section. In it you find the concepts of [*cache zone*](#cache-zones) and [*virtual host*](#virtual-hosts).

nedomi supports many virtual hosts and many cache zones. Every virtual host stores its cache in a single cache zone. But there may be many virtual hosts which store their cache in one zone.

The main sections of the config look like this.

```
{
    "system": {...},
    "cache_zones": [...],
    "http": {...},
    "logging": {...}
}
```

Descriptions of all keys and their values types follows.

* `system` - *object*, [read more](#system)
* `cache_zones` - *array* of *objects*, [read more](#cache-zones)
* `http` - *object*, [read more](#http-config)
* `logging` - *object*, [read more](#logging)

### HTTP Config

Here you can find all the HTTP-related configurations. The basic config looks like this:

```
{
    "listen": ":8282",
    "max_headers_size": 1231241212,
    "read_timeout": 12312310,
    "write_timeout": 213412314,
    "status_page": "/status",
    "virtual_hosts": [...]
}
```

Desciption of all the keys and their meaning:

* `listen` (*string*) - Sets the listening address and port of the server. Supports [golang's net.Dial addresses](http://golang.org/pkg/net/#Dial). Examples: `:80`, `example.com:http`, `192.168.133.25:9293`

* `max_headers_size` (*int*) - How much of a request headers (in **bytes**) will the server read before sending an error to the client.

* `read_timeout` (*int*) - Sets the reading timeout (in **seconds**) of the sever. If reading for a client takes this long the connection will be closed.

* `write_timeout` (*int*) - Similar to `read_timeout` but for writing the response. If the writing take too long the connection will be closed to.

* `status_page` (*string*) - The path of the server's [status page](#status-page). It musts start with a slash.

### Virtual Hosts

Virtual hosts are something familiar if you are coming form [apache](https://httpd.apache.org/docs/2.2/vhosts/). In nginx they are called [servers](http://wiki.nginx.org/HttpCoreModule#server). Basically you can have different behaviours depending on the `Host` header sent to your server.

### Cache Zones

Our Cache zones are very similar to the [nginx' cache zones](http://nginx.com/resources/admin-guide/caching/) in that they represent bounded space on the storage for a cache. If files stored in this space exeed its limitations the worst (cacheing-wise) files will be removed to get it back to the desired limits.

### System

### Logging

### Full Config Example

## Algorithms

## Status Page

## Limitations

* listen port
* upstreams
