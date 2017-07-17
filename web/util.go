package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/autograde/aguis/models"
	"github.com/labstack/echo"
)

// parseUint takes a string and returns the corresponding uint64. If the string
// parses to 0 or an error occurs, an error is returned.
func parseUint(s string) (uint64, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n == 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid identifier")
	}
	return n, nil
}

// parseStatuses takes a string of comma separated statuses and returns a slice
// of the corresponding status constants.
func parseStatuses(s string) ([]uint, bool) {
	if s == "" {
		return []uint{}, true
	}

	ss := strings.Split(s, ",")
	if len(ss) > 3 {
		return []uint{}, false
	}
	var statuses []uint
	for _, s := range ss {
		switch s {
		case "pending":
			statuses = append(statuses, models.Pending)
		case "rejected":
			statuses = append(statuses, models.Rejected)
		case "accepted":
			statuses = append(statuses, models.Accepted)
		default:
			return []uint{}, false
		}
	}
	return statuses, true
}
