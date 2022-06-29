# Middleware in Golang

This folder contains my notes and examples of how to write HTTP middleware in Go.

It uses the the standard `net/http` package, [`zap`](https://github.com/uber-go/zap) for logging, and [`Alice`](https://github.com/justinas/alice) for easier chaining of middleware.

References:

* https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702#.e4k81jxd3
* https://www.nicolasmerouze.com/middlewares-golang-best-practices-examples

## Middleware

Currently there are three middleware methods:

* `LogMiddleware` - logs a start request and end request message.
* `ResponseMesserMiddleware` - prepends a naughty note to the body of the response by intercepting the `ResponseWriter` object and redirecting its writes.
* `AddHeaderMiddleware` - adds an arbitrary header value to the response. This was interesting because I wanted to make the method generic while still making it compatible with Alice, so I actually had to write a method that returned the method expected by Alice's `New` method.

## Testing

All three middlewares have unit tests, which are written in the style of "black-box" tests (i.e. they are written using only the public interfaces of the package). See [this GitHub issue](https://github.com/golang/go/issues/25223) and the linked [StackOverflow answer](https://stackoverflow.com/questions/19998250/proper-package-naming-for-testing-with-the-go-language/31443271#31443271) for more details on this pattern.

Additionally, you can see some integration testing of the implementations of the middleware in the API in the [api/integration](../integration/) folder.
