package ci

import "fmt"

// Environment variable used by the CI system to pass
// the session secret from QuickFeed to the test code.
const secretEnvName = "QUICKFEED_SESSION_SECRET"

var ErrConflict = fmt.Errorf("submission is already being built, please wait")
