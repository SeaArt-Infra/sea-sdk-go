package sa

import "net/http"

// PassthroughResponse is the raw response returned by a vendor passthrough API.
type PassthroughResponse struct {
	StatusCode int
	Headers    http.Header
	Body       RawResponse
}
