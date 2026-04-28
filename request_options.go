package sa

import "net/http"

// RequestOption configures a single SDK request.
type RequestOption interface {
	apply(*requestOptions)
}

type requestOptions struct {
	headers http.Header
}

type requestOptionFunc func(*requestOptions)

func (f requestOptionFunc) apply(opts *requestOptions) {
	f(opts)
}

func buildRequestOptions(opts []RequestOption) requestOptions {
	var cfg requestOptions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt.apply(&cfg)
	}
	return cfg
}

// WithHeader sets one request header for a single SDK call.
func WithHeader(key, value string) RequestOption {
	return requestOptionFunc(func(opts *requestOptions) {
		if opts.headers == nil {
			opts.headers = make(http.Header)
		}
		opts.headers.Set(key, value)
	})
}

// WithHeaders sets request headers for a single SDK call.
func WithHeaders(headers http.Header) RequestOption {
	return requestOptionFunc(func(opts *requestOptions) {
		if len(headers) == 0 {
			return
		}
		if opts.headers == nil {
			opts.headers = make(http.Header, len(headers))
		}
		for key, values := range headers {
			opts.headers.Del(key)
			for _, value := range values {
				opts.headers.Add(key, value)
			}
		}
	})
}
