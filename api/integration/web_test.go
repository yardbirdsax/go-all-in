//go:build integration
package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWeb(t *testing.T) {
	URL := os.Getenv("WEB_URL")
	expectedVersion := os.Getenv("WEB_EXPECTED_VERSION")

	testCases := []struct{
		name string
		path string
		expectedHeaders map[string][]string
		expectedBody string
	}{
		{
			name: "rootPath",
			path: "?name=Josh",
			expectedHeaders: map[string][]string{},
			expectedBody: fmt.Sprintf("Hello %s, version %s", "Josh", expectedVersion),
		},
		{
			name: "withHeaders",
			path: "with-headers?name=Josh",
			expectedHeaders: map[string][]string{
				"X-My-Header": {"Value"},
			},
			expectedBody: fmt.Sprintf("Hello %s, version %s", "Josh", expectedVersion),
		},
	}

	for _, testCase := range(testCases) {
		t.Run(testCase.name, func(t *testing.T) {
			fullURL := fmt.Sprintf("%s/%s", URL, testCase.path)
			req, err := http.NewRequest("GET", fullURL, nil)
			require.Nil(t, err, "NewRequest returned an unexpected error")

			client := http.Client{}
			response, err := client.Do(req)
			require.NoError(t, err, "client.Do returned an unexpected error")

			for expectedHeaderKey, expectedHeaderValue := range(testCase.expectedHeaders) {
				assert.Contains(t, response.Header, expectedHeaderKey, "expected header key is missing")
				actualHeaderValue, ok := response.Header[expectedHeaderKey]
				if ok {
					assert.Equal(t, expectedHeaderValue, actualHeaderValue, "expected header value is not equal to actual header value")
				}
			}
			bodyBytes, err := ioutil.ReadAll(response.Body)
			require.Nil(t, err, "ReadAll body returned an unexpected error")
			body := string(bodyBytes)
			assert.Equal(t, testCase.expectedBody, body, "expected response body does not match actual body")
		})
	}
}
