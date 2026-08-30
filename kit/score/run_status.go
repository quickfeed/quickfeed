package score

// Failed reports whether the run did not complete successfully.
func (s RunStatus) Failed() bool {
	return s != RunStatus_SUCCESS
}

// ScoresValid reports whether the run produced scores that should be recorded.
// A build failure has trustworthy zero scores even though the run failed.
func (s RunStatus) ScoresValid() bool {
	return s == RunStatus_SUCCESS || s == RunStatus_BUILD_FAILURE
}

// Failed reports whether the build information describes a failed run.
func (b *BuildInfo) Failed() bool {
	return b.GetStatus().Failed()
}

// ScoresValid reports whether the build produced scores that should be recorded.
func (b *BuildInfo) ScoresValid() bool {
	return b.GetStatus().ScoresValid()
}

// Failed reports whether the results describe a failed run.
func (r *Results) Failed() bool {
	return r.GetBuildInfo().Failed()
}

// ScoresValid reports whether the results contain scores that should be recorded.
func (r *Results) ScoresValid() bool {
	return r.GetBuildInfo().ScoresValid()
}
