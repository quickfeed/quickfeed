package scm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

func MustParse[N ~int | ~int64](s string) N {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return N(i)
}

func MustUnmarshal[T any](r io.Reader) T {
	var v T
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		panic(err)
	}
	return v
}

func MustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// MockBackendOption is used to configure the *http.ServeMux for the mocked backend.
type MockBackendOption func(*http.ServeMux)

func WithRequestMatchHandler(pattern string, handler http.Handler) MockBackendOption {
	return func(router *http.ServeMux) {
		router.Handle(pattern, handler)
	}
}

// EnforceHostRoundTripper rewrites all requests with the given `Host`.
type EnforceHostRoundTripper struct {
	Host                 string
	UpstreamRoundTripper http.RoundTripper
}

// RoundTrip implementation of `http.RoundTripper`
func (rt *EnforceHostRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	splitHost := strings.Split(rt.Host, "://")
	r.URL.Scheme = splitHost[0]
	r.URL.Host = splitHost[1]
	return rt.UpstreamRoundTripper.RoundTrip(r)
}

func NewMockedHTTPClient(options ...MockBackendOption) *http.Client {
	router := http.NewServeMux()
	for _, o := range options {
		o(router)
	}
	mockServer := httptest.NewServer(router)
	c := mockServer.Client()
	c.Transport = &EnforceHostRoundTripper{
		Host:                 mockServer.URL,
		UpstreamRoundTripper: mockServer.Client().Transport,
	}
	return c
}
