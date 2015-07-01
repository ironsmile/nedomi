# Upstream Modules

This directory contains all upstream modules. An upstream module is one which is resposible for downloading content from a remote host via HTTP.

## Contents

* [Anatomy of a Upstream Module](#anatomy-of-a-upstream-module)
* [How to Write Your Own Module?](#how-to-write-your-own-module)
* [Removing a Module](#removing-a-module)
* [How Does it Work?](#how-does-it-work)


## Anatomy of a Upstream Module

It is a subpackage in the `upstream/` directory which must have a new function of the type:

```
func New() *T
```

where `T` is a type which implements the [Upstream](https://godoc.org/github.com/ironsmile/nedomi/upstream#Upstream) interface defined in [interface.go](interface.go).


## How to Write Your Own Module?

You can add your own modules as long as their names do not collide with any other module's name. Lets say you want to create a round robin upstream module which will get files from multiple origins. And you want your module be called `rrobin`.

* Go into the `upstream/` directory - `$ cd .../nedomi/upstream`
* Create a directory which will be the name of your module. Lets say it is **rrobin** so it is `mkdir rrobin`
* Write your implementation of [Upstream](https://godoc.org/github.com/ironsmile/nedomi/upstream#Upstream) in this directory as `package rrobin`
* In the main project directory run `go generate ./...`


## Removing a Module

Lets say you want to remove the *rrobin* module.

* Within the project directory
* `rm -rf upstream/rrobin`
* `go generate ./...`

You will have to leave at least one upstream module in order to have a working build.


## How Does it Work?

`golang` does not support dynamic linking. So our only choice is to come up with some other trick to have optional modules. For that reason we use the `go generate` tool.
