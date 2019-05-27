package web

import (
	"context"
	"fmt"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
)

// ListDirectories returns all directories which can be used as a course
// directory from the given provider.
func ListDirectories(ctx context.Context, scm scm.SCM) (*pb.Directories, error) {

	ctx, cancel := context.WithTimeout(ctx, MaxWait)
	defer cancel()

	directories, err := scm.ListDirectories(ctx)
	if err != nil {
		return nil, err
	}

	organizations := make([]*pb.Directory, 0)
	for _, directory := range directories {
		plan, err := scm.GetPaymentPlan(ctx, directory.ID)
		if err != nil {
			fmt.Println("Error getting payment plan for directory ID: ", directory.ID)
		}
		if plan.PrivateRepos > 0 {
			organizations = append(organizations, directory)
		}
	}

	return &pb.Directories{Directories: organizations}, nil
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
