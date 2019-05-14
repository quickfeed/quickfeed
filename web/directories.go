package web

import (
	"context"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
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

func ListDirectories(ctx context.Context, scm scm.SCM) (*pb.Directories, error) {

	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	directories, err := scm.ListDirectories(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.Directories{Directories: directories}, nil
}

/*
func ListDirectories() echo.HandlerFunc {
	return func(c echo.Context) error {

		log.Println("Listing directories: still REST")
		var dr ListDirectoriesRequest
		if err := c.Bind(&dr); err != nil {
			return err
		}
		if !dr.valid() {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
		}

		s, err := getSCM(c, dr.Provider)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), MaxWait)
		defer cancel()

		directories, err := s.ListDirectories(ctx)
		if err != nil {
			return err
		}

		return c.JSONPretty(http.StatusOK, directories, "\t")
	}
}*/
