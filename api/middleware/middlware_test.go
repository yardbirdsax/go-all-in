//go:build unit
package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yardbirdsax/go-all-in/api/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestResponseMesserMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "http://testing", nil)
	resp := httptest.NewRecorder()
	body := "Hello world!"
	expectedBody := fmt.Sprintf("I messed you up!\n%s", body)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	})
	
	testHandler := middleware.ResponseMesserMiddleware(handler)
	testHandler.ServeHTTP(resp, req)

	assert.Equal(t, expectedBody, resp.Body.String())
}

func TestLogMiddleware(t *testing.T) {
	path := "path"
	expectedPath := fmt.Sprintf("/%s", path)
	expectedHost := "host"
	expectedURL := fmt.Sprintf("http://%s/%s", expectedHost, path)
	req := httptest.NewRequest("GET", expectedURL, nil)
	resp := httptest.NewRecorder()
	observerCore, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(observerCore)
	zap.ReplaceGlobals(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testHandler := middleware.LogMiddleware(handler)
	testHandler.ServeHTTP(resp, req)

	assert.Equal(t, 2, observedLogs.Len())
	assert.Equal(t, 1, observedLogs.FilterMessageSnippet("Starting request").FilterField(zap.String("path", expectedPath)).Len())
	assert.Equal(t, 1, observedLogs.FilterMessageSnippet("Ending request").FilterField(zap.String("path", expectedPath)).Len())
}

func TestHeaderAdderMiddleware(t *testing.T) {
	expectedHeaderName := "X-My-Header"
	expectedHeaderValue := "value"
	req := httptest.NewRequest("GET", "http://test", nil)
	resp := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testHandler := middleware.HeaderAdderMiddleware(expectedHeaderName, expectedHeaderValue)(handler)
	testHandler.ServeHTTP(resp, req)

	assert.Contains(t, resp.Result().Header, expectedHeaderName)
	assert.Equal(t, expectedHeaderValue, resp.Header().Get(expectedHeaderName))
}
