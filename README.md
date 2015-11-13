# nedomi
[![Build Status](https://travis-ci.org/ironsmile/nedomi.svg)](https://travis-ci.org/ironsmile/nedomi)
[![Coverage Status](https://coveralls.io/repos/ironsmile/nedomi/badge.svg)](https://coveralls.io/r/ironsmile/nedomi)
[![GoDoc](https://godoc.org/github.com/ironsmile/nedomi?status.svg)](https://godoc.org/github.com/ironsmile/nedomi)

HTTP media cache server. Most caching servers does not understand when a media file is being proxied. When storing media files you can gain  performance and space when you consider the fact that most of the big meda files are not actually watched from end to end.

We intend to implement a caching algorithm which takes all this into consideration and delivers better cache performance and throughput.

## Contents

* [About](#nedomi)
* [Algorithms](#algorithms)
* [Requirements](#requirements)
* [Install](#install)
* [Configuration](#configuration)
* [Status Page](#status-page)
* [Benchmarks](#benchmarks)
* [Limitations](#limitations)
* [Extending It](#extending-it)
* [Credits](#credits)

## Algorithms

nedomi is designed so that we can change the way it works. For every major part of its internals it uses [interfaces](http://golang.org/doc/effective_go.html#interfaces). This will hopefully make it easier when swapping algorithms.

The most important one is the caching algorithm. At the moment nedomi has only one implemented and it is *segmented LRU*. It is inspired by [Varnish's idea](https://www.varnish-software.com/blog/introducing-varnish-massive-storage-engine). The big thing that makes it even better for nedomi is that our objects always have exactly the same size. We do not keep whole files in the cache but evenly sized parts of the files. This effectively means that the implementation of the cache evictions and insertions is extremely simple. It will be as easy to deal with storage fragmentation if we ever need to.

We keep track of file chunks separately. This means chunks that are not actually watched are not stored in the cache. Our observations in the real world show that when consuming digital media people more often than not skip parts and jump from place to place. Storing unwatched gigabytes does not make sense. And this is the real benefit of our chunked storage. It stores only the popular parts of the files which leads to better cache performance.


## Requirements

Nothing. It is pure Go. If you have some of [latest versions](https://golang.org/dl/) of the language you are good to go.

## Install

Nothing fancy. Just `go get`.

```sh
go get github.com/ironsmile/nedomi
```

At the moment this is the only way to install the software. In the future (when it gets more stable) we may start to distribute binary packages.

## Configuration

**Notice!** _This section is slightly out of date at this point. It will be cleaned up once the [alpha-1 milestone](https://github.com/ironsmile/nedomi/issues?q=is%3Aopen+is%3Aissue+milestone%3A%22Alpha+1%22) is reached. While mostly true not all of the things written here are correct. You are advised to look at the [config.example.json](config.example.json) file where everything should be up to date._

We all know that a mark of a good software is its configurability. Not everyone has the same needs. So we have you covered. nedomi lets you choose all the details.

The configuration is stored in a single [JSON file](https://en.wikipedia.org/wiki/JSON). It is basically an object with a key for every single section. In it you find the concepts of [*cache zone*](#cache-zones) and [*virtual host*](#virtual-hosts).

nedomi supports many virtual hosts and many cache zones. Every virtual host stores its cache in a single cache zone. But there may be many virtual hosts which store their cache in one zone.

The main sections of the config look like this.

```js
{
    "system": {/*...*/},
    "cache_zones": [/*...*/],
    "http": {/*...*/},
    "logger": {/*...*/}
}
```

Descriptions of all keys and their values types follows.

* `system` - *object*, [read more](#system)
* `cache_zones` - *array* of *objects*, [read more](#cache-zones)
* `http` - *object*, [read more](#http-config)
* `logging` - *object*, [read more](#logging)

If you feel want to just skip to the full example - [here is the place for your click](#full-config-example). For everyone else a description of all configuration options follows.

### HTTP Config

Here you can find all the HTTP-related configurations. The basic config looks like this:

```js
{
    "listen": ":8282",
    "max_headers_size": 1231241212,
    "read_timeout": 12312310,
    "write_timeout": 213412314,
    "io_tranfer_size": 128k,
    "virtual_hosts": [/*...*/],
}
```

Description of all the keys and their meaning:

* `listen` (*string*) - Sets the listening address and port of the server. Supports [golang's net.Dial addresses](http://golang.org/pkg/net/#Dial). Examples: `:80`, `example.com:http`, `192.168.133.25:9293`

* `max_headers_size` (*int*) - How much of a request headers (in **bytes**) will the server read before sending an error to the client.

* `read_timeout` (*int*) - Sets the reading timeout (in **seconds**) of the sever. If reading for a client takes this long the connection will be closed.

* `write_timeout` (*int*) - Similar to `read_timeout` but for writing the response. If the writing take too long the connection will be closed to.

* `virtual_hosts` (*array*) - Contains the [virtual hosts](#virtual-hosts) of this server. Every virtual host is represented by a object which contains its configuration.

* `io_transfer_size` (*string*) - Bytes size. It tells the size of blocks to be transfered on the network. The timeouts previously mentioned are for pieces this big. Too big of a size might lead to timing out or too excessive memory usage, too small may lead to bad performance due to too many syscalls. The default is '128k'.

### Cache Zones

Our Cache zones are very similar to the [nginx' cache zones](http://nginx.com/resources/admin-guide/caching/) in that they represent bounded space on the storage for a cache. If files stored in this space exceeds its limitations the worst (caching-wise) files will be removed to get it back to the desired limits.

Example cache zone:

```js
{
    "id": 2,
    "path": "/home/iron4o/playfield/nedomi/cache2",
    "storage_objects": 4723123,
    "part_size": "4m",
    "cache_algorithm": "lru",
    "skip_cache_key_in_path": true
}
```

* `id` (*int*) - unique ID of this cache zone. It will be used to match virtual hosts to cache zones.

* `path` (*string*) - path to a directory in which the cache for this zone will be stored.

* `storage_objects` (*int*) - the maximum amount of objects which will be stored in this cache zone. In conjunction with `part_size` they form the maximum disk space which this zone will take.

* `part_size` (*string*) - Bytes size. It tells on how big a chunks a file will be chopped when saved. It consists of a number and a size letter. Possible letters are 'k', 'm', 'g', 't' and 'z'. Sizes like "1g200m" are not supported at the moment, use "1200m" instead. This will probably change in the future.

* `cache_algorithm` (*string*) - Sets the cache eviction algorithm. You can see the possible algorithms in the `cache/` directory.

* `skip_cache_key_in_path` (*boolean*) - sets if the cache should be added as part of the path for each file in this cache zone. The default is false - add the cache key in front of the path for each cached file.

### Virtual Hosts

Virtual hosts are something familiar if you are coming form [apache](https://httpd.apache.org/docs/2.2/vhosts/). In nginx they are called [servers](http://wiki.nginx.org/HttpCoreModule#server). Basically you can have different behaviours depending on the `Host` header sent to your server.

Example virtual host:

```js
{
    "name": "proxied.example.com",
    "upstream": "http://example.com",
    "cache_zone": 2,
    "cache_key": "2.1"
}
```

* `name` (*string*) - The host name. It must match the `Host:` request header exactly. In that case the matching virtual host will be used.

* `upstream` (*string*) - Either a full HTTP/HTTPS address of the proxied server or the ID of an advanced upstream. 

* `cache_zone` (*int*) - ID of a cache zone in which files for this virtual host will be cached. It should match an id of defined cache zone.

* `cache_key` (*string*) - Key used for storing files in the cache. If two different virtual hosts share the same `cache_key` they will share their cache as well.

### System

All keys are:

```js
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

```js
{
    "log_file": "/var/log/nedomi.log",
    "debug": false
}
```

* `log_file` (*string*) - Path to the file in which all logs will be stored.
* `debug` (*boolean*) - If set to true the whole output will be send to stdout.

### Full Config Example

See the `config.example.json` in the repo. It is [here](config.example.json). It has all the possible keys and example values.

## Status Page

nedomi comes with a HTTP status page. It is a pretty html with information about the internals and performance of the running server. After you've determined the host (let's say `127.0.0.2`) you want to use for the status page, add the following record to the `virtual_hosts` section of the `http` configuration to enable the status page on that address:
```js
{
    "name": "127.0.0.2",
    "handler": "status"
}
```

## Benchmarks

Measuring performance with benchmarks is a hard job. We've tried to do it as best as possible. We used mainly [wrk](https://github.com/wg/wrk) for our benchmarks. Included in the repo is [one of our best scripts](tools/wrk_test.lua) and few [results form running it](benchmark-results) at various stages of the development.

This benchmark script tries to behave like a real users watching videos. It seeks from place to place, it likes some videos more than others. Also, it is more likely to watch the beginning of the video.

At the moment our measurements show that nedomi is comparable or slightly better than nginx in the tested work loads. We expect much more performance after code optimization which nedomi haven't had to this moment.

## Limitations

Needless to say this is a young project. There are things which we know are required from any good cache server but we simple haven't time to put in.

**Listening addresses** - At the moment nedomi only listens to one address. All virtual hosts must use it.

**Upstreams** - Only one upstream is supported for a virtual host.

**Cache Loading** - For the moment nedomi will not load its cache from the disk after restart.

**HTTPS** is not supported.

## Extending It

nedomi is modular. You can add or remove modules from it as much as you want. We've tried to abstarct away every possible thing into an interface. Extending normally is done via writing a package and adding it into its proper place in the source tree and then running `go generate ./...`. For specific instructions you should see the particular README for every type of modules which can be written. They are:

* types.CacheAlgorithm takes care of what responses should be cached and how many. See the [README.md](cache/) in `cache/`

* Storage which decides how exactly cached files will be stored. See [README.md](storage/) in `storage/`

* *Future:* Output filter modules in [modules/output](modules/output)

* *Future:* Input filter modules in [modules/input](modules/input)

## Credits

We've used videos from [OpenFest](http://www.openfest.org/) lectures for our benchmarks.

*There are 2 hard problems in computer science: caching, naming and off-by-1 errors* - Tim Bray quoting Phil Karlton and extended by the Internet
