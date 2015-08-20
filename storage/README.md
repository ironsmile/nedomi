# Storage Modules

The logic for storing cached files in nedomi is highly modular. At the moment we have a built in storage on disk. But you can have as many and as different as you want. They are all subpackages in the `storage/` directory.

## Contents

* [Anatomy of a Storage Module](#anatomy-of-a-storage-module)
* [How to Write Your Own Module?](#how-to-write-your-own-module)
* [Removing a Module](#removing-a-module)

## Anatomy of a Storage Module

It is a subpackage in the `storage/` directory. It follows the follwing rules:

* Has a `func New(cfg config.CacheZoneSection, ca types.CacheAlgorithm, logger types.Logger) *T` function.

* `T` must conform to the Storage interface which is defined in [storage/interface.go](interface.go)

## How to Write Your Own Module?

You can add your module by creating a directory with a subpackage in the `storage/` directory.

* Go into the `storage/` directory with `$ cd .../nedomi/storage`
* Create a directory for your module. The name of the directory will be name of the module. Lets say you want to create a redis module. `mkdir redis`
* Write your implementation of Storage in this directory as `package redis`
* In the main nedomi directory run `cd .../nedomi && go generate ./...`

## Removing a Module

Lets say you want to remove the *redis* module.

* `cd .../nedomi/cache`
* `rm -rf redis`
* `cd .. && go generate ./...`

You can remove any storage module as well. Including the built in modules. Just make sure there is at least one left. Otherwise you wouldn't be able to start the server after compiling. The source will compile happily without any modules left, though.
