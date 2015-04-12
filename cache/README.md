# Caching Modules

nedomi is very modular when it comes to the cache replacing algorithms. You can have different caching methods for every cache zone. On top of that you can write your own cache replacing algorithm if the built in are not enough for you. Every algorithm is a package which conforms to few rules which will be described below. Even the built in algorithms for caching replacement are implemente in this way. They can be used for example. Or they can be removed from your build whatsoever.

*module* or *cache module* will be used from here on instead of the much more descriptive "cache replacing algorithm package".


## Contents

* [Anatomy of a Cache Module](#anatomy-of-a-cache-module)
* [How to Write Your Own Module?](#how-to-write-your-own-module)
* [Removing a Module](#removing-a-module)
* [How Does it Work?](#how-does-it-work)


## Anatomy of a Cache Module

It is a subpackage in the `cache/` directory which does to following:

* Has a `New (cz *config.CacheZoneSection) *T` function where `config` is `github.com/ironsmile/nedomi/config`.
* `T` must conform to the `CacheManager` interface which is defined in [cache/interface.go](interface.go).


## How to Write Your Own Module?

You can add your own modules as long as their name does not collide with any other caching module's name.

* Go into the `cache/` directory - `$ cd .../nedomi/cache`
* Create a directory which will be the name of your module. Lets say it is **random** so it is `mkdir random`
* Write your implementation of CacheManager in this directory as `package random`
* In the main nedomi directory run `cd .../nedomi && go generate ./...`


## Removing a Module

Lets say you want to remove the *random* module.

* `cd .../nedomi/cache`
* `rm -rf random`
* `cd .. && go generate ./...`

You can remove any caching module as well. Including the built in modules. Just make sure there is at least one left. Otherwise you wouldn't be able to start the server after compiling. The source will compile happily without any modules left, though. 


## How Does it Work?

`golang` does not support dynamic linking. So our only choice is to come up with some other trick to have optional modules. We use the `go generate` for this.

At the moment the only way a CacheManager is created in nedomi is through the [NewCacheManager](new_cache_manager.go) function. It looks for CacheManager implementation in the [cacheTypes](types.go) map. This map is generated via `go generate` which in turn uses the [generate_cache_types](generate_cache_types) script.
