<div align="center">

![Flow](https://raw.githubusercontent.com/alexedwards/flow/assets/flow-sm.png)
        
[![Go Reference](https://pkg.go.dev/badge/github.com/alexedwards/flow.svg)](https://pkg.go.dev/github.com/alexedwards/flow) [![Go Report Card](https://goreportcard.com/badge/github.com/alexedwards/flow)](https://goreportcard.com/report/github.com/alexedwards/flow) [![MIT](https://img.shields.io/github/license/alexedwards/flow)](https://img.shields.io/github/license/alexedwards/flow) ![Code size](https://img.shields.io/github/languages/code-size/alexedwards/flow)

A delightfully tiny but powerful HTTP router for Go web applications
</div>

---

Flow packs in a bunch of features that you'll probably like:

* Use **named parameters**, **wildcards** and (optionally) **regexp patterns** in your routes.
* Create route **groups which use different middleware** (a bit like chi).
* **Customizable handlers** for `404 Not Found` and `405 Method Not Allowed` responses.
* **Automatic handling** of `OPTIONS` and `HEAD` requests.
* Works with `http.Handler`, `http.HandlerFunc`, and standard Go middleware.
* Zero dependencies.
* Tiny, readable, codebase (~160 lines of code).

---

### Installation

```
$ go get github.com/alexedwards/flow@latest
```

### Basic example

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/alexedwards/flow"
)

func main() {
    // Initialize a new router.
    mux := flow.New()

    // Add a `GET /greet/:name` route. The : character is used to denote a 
    // named parameter in the URL path, which acts like a 'wildcard'.
    mux.HandleFunc("/greet/:name", greet, "GET")

    err := http.ListenAndServe(":2323", mux)
    log.Fatal(err)
}

func greet(w http.ResponseWriter, r *http.Request) {
    // Use flow.Param() to retrieve the value of the named parameter from the
    // request context.
    name := flow.Param(r.Context(), "name")

    fmt.Fprintf(w, "Hello %s", name)
}
```

### Kitchen-sink example

```go
mux := flow.New()

// The Use() method can be used to register middleware. Middleware declared at
// the top level will used on all routes (including error handlers and OPTIONS
// responses).
mux.Use(exampleMiddleware1)

// Routes can use multiple HTTP methods.
mux.HandleFunc("/profile/:name", exampleHandlerFunc1, "GET", "POST")

// Optionally, regular expressions can be used to enforce a specific pattern
// for a named parameter.
mux.HandleFunc("/profile/:name/:age|^[0-9]{1,3}$", exampleHandlerFunc2, "GET")

// The wildcard ... can be used to match the remainder of a request path.
// Notice that HTTP methods are also optional (if not provided, all HTTP
// methods will match the route). The value of the wildcard can be retrieved 
// by calling flow.Param("...").
mux.Handle("/static/...", exampleHandler)

// You can create route 'groups'.
mux.Group(func(mux *flow.Mux) {
    // Middleware declared within in the group will only be used on the routes
    // in the group.
    mux.Use(exampleMiddleware2)

    mux.HandleFunc("/admin", exampleHandlerFunc3, "GET")

    // Groups can be nested.
    mux.Group(func(mux *flow.Mux) {
        mux.Use(exampleMiddleware3)

        mux.HandleFunc("/admin/passwords", exampleHandlerFunc4, "GET")
    })
})
```

### Notes

* Conflicting routes are permitted (e.g. `/posts/:id` and `posts/new`). Routes are matched in the order that they are declared.
* Trailing slashes are significant (`/profile/:id` and `/profile/:id/` are not the same).
* An `Allow` header is automatically set for all `OPTIONS` and `405 Method Not Allowed` responses (including when using custom handlers). 
* Once the `flow.Mux` type is being used by your server, it is *not safe* to add more middleware or routes concurrently.
* Middleware must be declared *before* a route in order to be used by that route. Any middleware declared after a route won't act on that route. For example:

```go
mux := flow.New()
mux.Use(middleware1)
mux.HandleFunc("/foo", ...) // This route will use middleware1 only.
mux.Use(middleware2)
mux.HandleFunc("/bar", ...) // This route will use both middleware1 and middleware2.
```

### Contributing

Bug fixes and documentation improvements are very welcome! For feature additions or behavioral changes, please open an issue to discuss the change before submitting a PR.

### Thanks

The pattern matching logic for Flow was heavily inspired by [matryer/way](https://github.com/matryer/way).
