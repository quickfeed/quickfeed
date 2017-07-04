package web

import (
	"context"
	"net/http"

	"github.com/autograde/aguis/scm"
	"github.com/labstack/echo"
)

// ListDirectoriesRequest represents a request to list all directories for a
// given provider.
type ListDirectoriesRequest struct {
	Provider string `json:"provider"`
}

func (dr *ListDirectoriesRequest) valid() bool {
	return dr != nil && dr.Provider != ""
}

// ListDirectories returns all directories which can be used as a course
// directory from the given provider.
func ListDirectories() echo.HandlerFunc {
	return func(c echo.Context) error {
		var dr ListDirectoriesRequest
		if err := c.Bind(&dr); err != nil {
			return err
		}
		if !dr.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		if c.Get(dr.Provider) == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "provider "+dr.Provider+" not registered")
		}
		s := c.Get(dr.Provider).(scm.SCM)

		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		directories, err := s.ListDirectories(ctx)
		if err != nil {
			return err
		}

		return c.JSONPretty(http.StatusOK, directories, "\t")
	}
}
