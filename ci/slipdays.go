package ci

import "time"

func slipdays(rData *RunData) {
	//TODO(meling) should start be passed in?
	start := time.Now()
	// TODO(meling) move this code to separate file and function: slipdays.go
	// TODO(meling) write tests for this stuff first
	//TODO(meling) must also handle groups; if group assignment is late; withdraw one slip day per group member.
	timeUntilDeadline := rData.Assignment.DurationUntilDeadline(start)
	if timeUntilDeadline < 0 {
		// TODO(meling) implement steps below
		// get course slip days from database
		// get user's current slip day count from database
		// check whether the user has already exceeded or will exceed the course maximum
		// update slip days for user in database
		rData.Repo.GetUserID()
	}

	//TODO(meling) also need to propogate slip day information to frontend to show to user
	//TODO(meling) database schema will be updated
}
