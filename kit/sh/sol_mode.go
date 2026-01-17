package sh

const SkipMessage = "Skipping test when running without race detector"

func RunningWithSolution() bool {
	return SolutionTag == "solution"
}
