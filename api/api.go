package api

import (
	"fmt"
	"net/http"
)

type Client struct {
	version string
}

func NewClient(version string) (*Client, error){
	client := &Client{
		version: version,
	}
	return client, nil
}

func (c *Client) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if name := r.URL.Query().Get("name"); name != "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello %s, version %s", name, c.version)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}