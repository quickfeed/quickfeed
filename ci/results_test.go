package ci

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestExtractResult(t *testing.T) {
	out := `here is some output in the log.

{"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	res, err := ExtractResult(zap.NewNop(), out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.BuildInfo.BuildLog, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73") {
		t.Fatal("build log contains secret")
		t.Logf("res %+v", res.BuildInfo)
	}
}

func TestExtractResultWithWhitespace(t *testing.T) {
	out := `here is some output in the log with whitespace before the JSON string below.

    {"Secret":"59fd5fe1c4f741604c1beeab875b9c789d2a7c73","TestName":"Gradle","Score":100,"MaxScore":100,"Weight":1}

Here are some more logs for the student.
`

	res, err := ExtractResult(zap.NewNop(), out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.BuildInfo.BuildLog, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73") {
		t.Fatal("build log contains secret")
		t.Logf("res %+v", res.BuildInfo)
	}
}
