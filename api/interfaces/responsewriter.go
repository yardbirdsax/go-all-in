package interfaces

import (
	"net/http"
)

// Type ResponseWriter is used to mimic a `http.ResponseWriter` interface. It's used in writing middlewares
// that intercept calls to the ResponseWriter struct, enabling things like capturing body writes and status code
// sets.
type ResponseWriter interface {
	http.ResponseWriter
}