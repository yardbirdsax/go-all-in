package main

import (
	"os"
	"net/http"

	"github.com/justinas/alice"
	"github.com/yardbirdsax/go-all-in/api"
	"github.com/yardbirdsax/go-all-in/api/middleware"

	"go.uber.org/zap"
)

var version string

func main() {

	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.L().Panic(err.Error())
	}
	zap.ReplaceGlobals(logger)

	version = os.Getenv("API_VERSION")
	if version == "" {
		version = "DEV"
	}
	
	// Alice lets us easily chain together middlewares without making a really messy function chain
	commonHandlers := alice.New(
		middleware.LogMiddleware,
	)

	apiClient, err := api.NewClient(version)
	if err != nil {
		zap.L().Panic(err.Error())
	}

	http.Handle("/", commonHandlers.ThenFunc(apiClient.HandleRequest))

	withHeaderHandlers := alice.New(
		commonHandlers.Then,
		middleware.HeaderAdderMiddleware("X-My-Header", "Value"),
	)

	http.Handle("/with-headers", withHeaderHandlers.ThenFunc(apiClient.HandleRequest))

	withMesserHandlers := alice.New(
		commonHandlers.Then,
		middleware.ResponseMesserMiddleware,
	)

	http.Handle("/with-messer", withMesserHandlers.ThenFunc(apiClient.HandleRequest))

	http.ListenAndServe(":8080", nil)
}
