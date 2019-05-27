package web

import (
	"net/http"
	"strconv"
	"strings"

	pb "github.com/autograde/aguis/ag"
	"github.com/labstack/echo"
)

// GetEventsURL returns the event URL for a given base URL and a provider.
func GetEventsURL(baseURL, provider string) string {
	return GetProviderURL(baseURL, "hook", provider, "events")
}

// GetProviderURL returns a URL endpoint given a base URL and a provider.
func GetProviderURL(baseURL, route, provider, endpoint string) string {
	return "https://" + baseURL + "/" + route + "/" + provider + "/" + endpoint
}

// parseUint takes a string and returns the corresponding uint64. If the string
// parses to 0 or an error occurs, an error is returned.
func parseUint(s string) (uint64, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n == 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid identifier")
	}
	return n, nil
}

var errInvalidStatus = echo.NewHTTPError(http.StatusBadRequest, "invalid status query")

// parseEnrollmentStatus takes a string of comma separated status values
// and returns a slice of the corresponding status constants.
func parseEnrollmentStatus(s string) ([]uint, error) {
	if s == "" {
		return []uint{}, nil
	}

	ss := strings.Split(s, ",")
	if len(ss) > 4 {
		return []uint{}, errInvalidStatus
	}
	var statuses []uint
	for _, s := range ss {
		switch s {
		case "pending":
			statuses = append(statuses, uint(pb.Enrollment_Pending))
		case "rejected":
			statuses = append(statuses, uint(pb.Enrollment_Rejected))
		case "student":
			statuses = append(statuses, uint(pb.Enrollment_Student))
		case "teacher":
			statuses = append(statuses, uint(pb.Enrollment_Teacher))
		default:
			return []uint{}, errInvalidStatus
		}
	}
	return statuses, nil
}
