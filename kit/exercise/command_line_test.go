package exercise

import (
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

func TestCmdLine(t *testing.T) {
	t.Skip("This is expected to fail, so we skip it when running normally (see comment).")

	// TODO(meling) the following works, but doesn't exercise the test failure
	// cmds := Commands{
	// 	{Command: "ls -l", Result: "command_line.go", Search: ResultContains},
	// 	{Command: "ls -a", Result: "command_line.go", Search: ResultContains},
	// }
	// The following is expected to fail:
	cmds := Commands{
		{Command: "ls -l", Result: "command_line.go", Search: ResultDoesNotContain},
		{Command: "ls -a", Result: "command_line.go", Search: ResultDoesNotContain},
		{Command: "obviouslyDoesNotWork", Result: "works", Search: ResultEquals},
	}
	sc := score.NewScoreMax(t, 10, 1)
	outs := CommandLine(t, sc, cmds)
	for i := 0; i < len(cmds); i++ {
		t.Logf("stdout: %s", outs[i].StdOut())
		t.Logf("stderr: %s", outs[i].StdErr())
		if outs[i].err == nil {
			t.Logf("STEP %d: ERROR IS NIL", i)
		}
		t.Logf("err: %v", outs[i].Error())
	}
}
