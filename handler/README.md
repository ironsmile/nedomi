# Handler Modules

This directory contains all handler modules. A handler module is one which is defines the main logic for a request response. `nedomi` is very modular in this respect and you are welcome to write you own handler modules. Every subdirectory in here is a handler module.

A virtual host must have exactly one handler module and it is the brain of the request handling.

## Contents

* [Anatomy of a Cache Module](#anatomy-of-a-handler-module)
* [How to Write Your Own Module?](#how-to-write-your-own-module)
* [Removing a Module](#removing-a-module)
* [How Does it Work?](#how-does-it-work)


## Anatomy of a Handler Module

It is a subpackage in the `handler/` directory which must have a new function of the type:

```
func New() *T
```

where `T` is a type which implements the [RequestHandler](https://godoc.org/github.com/ironsmile/nedomi/handler#RequestHandler) interface defined in [interface.go](interface.go).


## How to Write Your Own Module?

You can add your own modules as long as their name does not collide with any other caching module's name.

* Go into the `handler/` directory - `$ cd .../nedomi/handler`
* Create a directory which will be the name of your module. Lets say it is **random** so it is `mkdir random`
* Write your implementation of RequestHandler in this directory as `package random`
* In the main project directory run `go generate ./...`


## Removing a Module

Lets say you want to remove the *random* module.

* Within the project directory
* `rm -rf handler/random`
* `go generate ./...`

You will have to leave at least one handler module in order to have a working build.


## How Does it Work?

`golang` does not support dynamic linking. So our only choice is to come up with some other trick to have optional modules. For that reason we use the `go generate` tool.
