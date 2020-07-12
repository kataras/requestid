# Request ID

[![build status](https://img.shields.io/travis/com/kataras/requestid/master.svg?style=for-the-badge&logo=travis)](https://travis-ci.com/github/kataras/requestid) [![report card](https://img.shields.io/badge/report%20card-a%2B-ff3333.svg?style=for-the-badge)](https://goreportcard.com/report/github.com/kataras/requestid) [![godocs](https://img.shields.io/badge/go-%20docs-488AC7.svg?style=for-the-badge)](https://godoc.org/github.com/kataras/requestid)

Unique Identifier for each HTTP request. Useful for logging, propagation and e.t.c.

## Installation

The only requirement is the [Go Programming Language](https://golang.org/dl).

```sh
$ go get github.com/kataras/requestid
```

## Getting Started

Import the package:

```go
package main

import "github.com/kataras/requestid"
```

Wrap a handler with the `Handler` function and retrieve the request ID using the `Get` function:

```go
import "net/http"

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        id:= requestid.Get(r)
        w.Write([]byte(id))
    })

    http.ListenAndServe(":8080", requestid.Handler(mux))
}
```

By-default the `requestid` middleware uses the `X-Request-Id` header to extract and set the request ID.
It generates a [universally unique identifier](https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_(random)) when the request header is missing. Use custom logic to extract and set the request ID using `HandlerWithGenerator`:

```go
import "net/http"

func main() {
    // extract from a request header and set to the response header.
    gen := func(w http.ResponseWriter, r *http.Request) string {
        id:= r.Header.Get("X-Custom-Id")
        if id == "" {
            // [custom logic to generate ID...]
        }
        w.Header().Set("X-Custom-Id", id)
        return id
    }

    // [...]
    router := requestid.HandlerWithGenerator(mux, gen)
    http.ListenAndServe(":8080", router)
}
```

## License

This software is licensed under the [MIT License](LICENSE).
