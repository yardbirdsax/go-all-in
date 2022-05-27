# Middleware in Golang

This folder contains my notes and examples of how to write HTTP middleware in Go.

It uses the the standard `net/http` package, [`zap`](https://github.com/uber-go/zap) for logging, and [`Alice`](https://github.com/justinas/alice for easier chaining of middleware.

References:

* https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702#.e4k81jxd3
* https://www.nicolasmerouze.com/middlewares-golang-best-practices-examples

## Basic web server functionality

* Accepts HTTP requests on port 8080, on the path "/". If another path is specified, it will return a 500 error.
* If a URL query parameter named "name" is specified, it will return a body that includes that name in a "Hello" style message.
* If the request doesn't include that, then it will return a 500 error.

## Middleware

All middleware code is in the [`middleware`](middleware/) folder.

Currently there are three middleware methods:

* `LogMiddleware` - logs a start request and end request message.
* `ResponseMesserMiddleware` - prepends a naughty note to the body of the response by intercepting the `ResponseWriter` object and redirecting its writes.
* `AddHeaderMiddleware` - adds an arbitrary header value to the response. This was interesting because I wanted to make the method generic while still making it compatible with Alice, so I actually had to write a method that returned the method expected by Alice's `New` method.