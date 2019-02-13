package web

import (
	"net/http"

	"github.com/autograde/aguis/scm"
	"github.com/labstack/echo"
)

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
