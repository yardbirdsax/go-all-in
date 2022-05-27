package main

import (
	"fmt"
	"net/http"

	"github.com/justinas/alice"
	"github.com/yardbirdsax/go-all-in/middleware/middleware"

	"go.uber.org/zap"
)

func main() {

	// Alice lets us easily chain together middlewares without making a really messy function chain
	commonHandlers := alice.New(
		middleware.LogMiddleware,
		middleware.ResponseMesserMiddleware,
		middleware.HeaderAdderMiddleware("X-My-Header", "Some-Value"),
		middleware.HeaderAdderMiddleware("X-My-Other-Header", "Some-Other-Value"),
	)

	http.Handle("/", commonHandlers.ThenFunc(handleRequest))

	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.L().Panic(err.Error())
	}
	zap.ReplaceGlobals(logger)

	http.ListenAndServe(":8080", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if name := r.URL.Query().Get("name"); name != "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello %s", name)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
