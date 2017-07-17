package web

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

// ParseUintParam returns the uint of the provided string s.
// If the length of s is 0, then 0 is returned.
func ParseUintParam(s string) (uint64, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n == 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid identifier")
	}
	return n, nil
}
