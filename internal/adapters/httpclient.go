package adapters

import (
	"net/http"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: defaultHTTPTimeout}
}
