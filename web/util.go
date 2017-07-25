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

var enrollStatusError = echo.NewHTTPError(http.StatusBadRequest, "invalid status query")

// parseEnrollmentStatus takes a string of comma separated status values
// and returns a slice of the corresponding status constants.
func parseEnrollmentStatus(s string) ([]uint, error) {
	if s == "" {
		return []uint{}, nil
	}

	ss := strings.Split(s, ",")
	if len(ss) > 3 {
		return []uint{}, enrollStatusError
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
			return []uint{}, enrollStatusError
		}
	}
	return statuses, nil
}
