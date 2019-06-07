package web

import (
	"net/http"
	"strings"

	pb "github.com/autograde/aguis/ag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/autograde/aguis/scm"
	"github.com/labstack/echo"
	"github.com/markbates/goth"
)

// TeacherSuffix is used to set user as a teacher to be able to create new course
const TeacherSuffix = "-teacher"

// getSCM is a helper to get the scm provider for the current echo context.
// Each user's context store information about their individual scm provider
// object. (see main.go and web/auth/auth.go for the code that registers the
// scm.SCM instance.)
func getSCM(c echo.Context, scmProvider string) (scm.SCM, error) {
	provider := c.Get(scmProvider)
	if provider == nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "provider "+scmProvider+" not registered")
	}
	// If type assertions fails, the recover middleware will catch the panic and log a stack trace.
	return provider.(scm.SCM), nil
}

// GetProviders returns a list of all providers enabled by goth
func GetProviders() (*pb.Providers, error) {
	var providers []string
	for _, provider := range goth.GetProviders() {
		if !strings.HasSuffix(provider.Name(), TeacherSuffix) {
			providers = append(providers, provider.Name())
		}
	}
	if len(providers) == 0 {
		return nil, status.Errorf(codes.NotFound, "no providers found")
	}
	return &pb.Providers{Providers: providers}, nil
}
