package scm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

func mustParse[N ~int | ~int64](s string) N {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return N(i)
}

func mustRead[T any](r io.Reader) T {
	var v T
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		panic(err)
	}
	return v
}

func mustWrite(w http.ResponseWriter, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	if _, err := w.Write(b); err != nil {
		panic(err)
	}
}

// replaceArgs replaces placeholders in the API pattern with provided values.
// Placeholders are expected to be in the format `{name}`, and the values are
// expected to be in the same order as the placeholders.
// It panics if the number of placeholders does not match the number of values,
// or if the placeholder format is invalid, such as missing the closing brace.
func replaceArgs(pattern string, args ...any) string {
	placeholders := strings.Count(pattern, "{")
	if placeholders != len(args) {
		panic(fmt.Sprintf("expected %d arguments, but got %d", placeholders, len(args)))
	}

	// Replace each placeholder with the corresponding argument
	for _, arg := range args {
		start := strings.Index(pattern, "{")
		if start == -1 {
			break
		}
		end := strings.Index(pattern, "}")
		if end == -1 || end < start {
			panic("invalid placeholder format")
		}
		pattern = fmt.Sprintf("%s%s=%v%s", pattern[:start], pattern[start+1:end], arg, pattern[end+1:])
	}
	return pattern
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
