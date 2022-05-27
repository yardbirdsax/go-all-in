package middleware

import (
	"go.uber.org/zap"
	"net/http"
)

// Log some stuff
func LogMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("Starting request", zap.String("path", r.URL.Path))
		h.ServeHTTP(w, r)
		zap.L().Info("Ending requst", zap.String("path", r.URL.Path))
	})
}

// Do some things to the response.
// There is probably a better way to do this besides re-implementing ResponseWriter!
func ResponseMesserMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("Starting response messer")
		interceptor := NewWriterIncerceptor(w)
		h.ServeHTTP(interceptor, r)
		body := "I messed you up!\n" + string(interceptor.Body)
		w.Write([]byte(body))
		zap.L().Info("Ending response messer")
	})
}

// This lets us add an arbitrary header key and value.
// To use this with Alice we have to have the method actually return the func that gets passed to Alice.
func HeaderAdderMiddleware(name string, value string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(name, value)
			h.ServeHTTP(w, r)
		})
	}
}

type WriterInterceptor struct {
	Body []byte
	R    http.ResponseWriter
}

func NewWriterIncerceptor(w http.ResponseWriter) *WriterInterceptor {
	return &WriterInterceptor{
		Body: []byte{},
		R:    w,
	}
}

func (w *WriterInterceptor) Write(b []byte) (int, error) {
	w.Body = append(w.Body, b...)
	return len(b), nil
}

func (w *WriterInterceptor) Header() http.Header {
	return w.R.Header()
}

func (w *WriterInterceptor) WriteHeader(statusCode int) {
	w.R.WriteHeader(statusCode)
}
