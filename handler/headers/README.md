# Headers - rewriter of headers

This handler enables changing both the response or request headers before calling the next handler.

The request headers are rewritten when it is the handler's turn in the handler chain. Rewriting the response headers is done at the handle's turn *and once more* before they are send to the client. This is to ensure that they have not been overwritten by the next handler. The rewriting of the response Headers before continuing is to enable the handlers in chain after `headers` to use the rewritten response headers - probably not a good practice, but it might be desirable.

All the headers keys will be canonized before processing. The settings should have "response" and "request" keys with values the same as config.HeadersRewrite (see the example below). If one is missing no action will be taken for the corresponding headers.


The order of execution is remove, add and then set - so the following config.RewriteHeaders:

```json
    "add_headers": {
        "via": ["nedomi", "20"],
        "comment": "unicorns"
    },
    "set_headers": {
        "via": "ned"
    },
    "remove_headers": ["comment"]
```

Will remove any "comment" headers, then add "via" and "comment" and then set the "via" header practically invalidating the add). Technically this is equal to:

```json
    "set_headers": {
        "comment": "unicorns",
        "via": "ned"
    }
```

As the example above shows `remove_headers`, `add_headers` and `set_headers` settings can take both a single string for value or an array of strings as values.

## Usage:

This handler can be configured like a normal handler:

```json
{

"handlers": [
    {
        "type": "headers",
        "settings": {
            "request": {
                "add_headers": {
                    "X-Header-Added": "via handler"
                }
            }
        }
    },
    {
        "type": "dir",
        "setttings": {}
    }
]

}
```

or using a shortcut configuration directly in the `http`, `vhost` or `location` blocks:

```json
{
    "virtual_hosts": {
        "localhost": {

            "add_headers": {
                "X-Header-Added": "via handler"
            },

            "locations": {
                "/some-place": {
                    "remove_headers": ["server"]
                }
            }
        }
    }
}
```

The two headers configurations above are identical. When used directly the `add_headers`, `remove_headers` and `set_headers` are executed in the context of the request headers. If you want to rewrite response headers, you will have to configure it as a handler.

## Handler Settings:

The following JSON:

```json
    {
        "response": {
            "add_headers": {
                "via": ["nedomi 0.1"]
            },
            "set_headers": {
                "Custom-Header": "a value"
            }
        },
        "request": {
            "remove_headers": ["pragma", "cache-control"]
        }
    }
```

Will add "via" header to the response and set the "Custom-Header". As well as remove the two headers "Pragma" and "Cache-Control" from the request.
