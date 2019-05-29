package web

import (
	"net/http"
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
