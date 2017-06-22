package web

import (
	"fmt"
	"net/http"
)

// HTTPError is a convenience function for writing a HTTP response given a
// status code and possibly an error.
func HTTPError(w http.ResponseWriter, code int, err error) {
	res := http.StatusText(code)
	if err != nil && debug {
		res = fmt.Sprintf("%s: %s", http.StatusText(code), err.Error())
	}
	http.Error(w, res, code)
}
