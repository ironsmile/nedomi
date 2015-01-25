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

* `status_page` (*string*) - The URI of the server's [status page](#status-page). It must start with a slash.

* `virtual_hosts` (*array*) - Contains the [virtual hosts](#virtual-hosts) of this server. Every virtual host is represented by a object which contains its configuration.

### Cache Zones

Our Cache zones are very similar to the [nginx' cache zones](http://nginx.com/resources/admin-guide/caching/) in that they represent bounded space on the storage for a cache. If files stored in this space exeed its limitations the worst (cacheing-wise) files will be removed to get it back to the desired limits.

Example cache zone:

```
{
    "id": 2,
    "path": "/home/iron4o/playfield/nedomi/cache2",
    "storage_objects": 4723123,
    "part_size": "4m"
}
```

* `id` (*int*) - unique ID of this cache zone. It will be used to match virtual hosts to cache zones.

* `path` (*string*) - path to a directory in which the cache for this zone will be stored.

* `storage_objects` (*int*) - the maximum amount of objects which will be stored in this cache zone. In conjuction with `part_size` they form the maximum disk space which this zone will take.

* `part_size` (*string*) - Bytes size. It tells on how big a chunks a file will be chopped when saved. It consists of a number and a size letter. Possible letters are 'k', 'm', 'g', 't' and 'z'. Sizes like "1g200m" are not supported at the moment, use "1200m" instead. This will probably change in the future.

### Virtual Hosts

Virtual hosts are something familiar if you are coming form [apache](https://httpd.apache.org/docs/2.2/vhosts/). In nginx they are called [servers](http://wiki.nginx.org/HttpCoreModule#server). Basically you can have different behaviours depending on the `Host` header sent to your server.

Example virtual host:

```
{
    "name": "proxied.example.com",
    "upstream_address": "http://example.com",
    "cache_zone": 2,
    "cache_key": "2.1"
}
```

* `name` (*string*) - The host name. It must match the `Host:` request header exactly. In that case the matching virtual host will be used.

* `upstream_address` (*string*) - HTTP address of the proxied server. It must contain the protocol. So "http://" or "https://" are required at the biginning of the address.

* `cache_zone` (*int*) - ID of a cache zone in which files for this virtual host will be cached. It should match an id of defined cache zone.

* `cache_key` (*string*) - Key used for storing files in the cache. If two different virtual hosts share the same `cache_key` they will share their cache as well.

### System

All keys are:

```
{
    "pidfile": "/tmp/nedomi_pidfile.pid",
    "workdir": "/",
    "user": "www-data"
}
```

* `pidfile` (*string*) - File path. nedomi will store its process ID in this file.
* `workdir` (*string*) - nedomi will set its working dir to this one on startup. This is handy for debugging and developing. When a coredump is created it will be in this directory.
* `user` (*string*) - Valid system user. nedomi will try to setuid to this user. Make sure the user which launches the binary has permissions for this.


### Logging

At the moment nedomi supports a single log file. All errors, notices (and even the access log) go in it. When the *-D* command line flag is used all output is send to the stdout.

```
{
    "log_file": "/var/log/nedomi.log",
    "debug": false
}
```

* `log_file` (*string*) - Path to the file in which all logs will be stored.
* `debug` (*boolean*) - If set to true the whole output will be send to stdout.

### Full Config Example

See the `config.example.json` in the repo. It is [here](config.example.json).

## Algorithms

## Status Page

## Limitations

* listen port
* upstreams
* not loading cache from disk
