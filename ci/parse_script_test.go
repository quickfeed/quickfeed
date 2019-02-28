package ci

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseScript(t *testing.T) {
	getURL := "https://github.com/uis-dat520-s2019/assignments.git"
	testURL := "https://github.com/uis-dat520-s2019/tests.git"
	info := AssignmentInfo{
		AssignmentName:     "lab2",
		Language:           "go",
		CreatorAccessToken: "secret",
		GetURL:             getURL,
		TestURL:            testURL,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(getURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(testURL, ".git"), "https://"),
	}
	job, err := ParseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("im:", job.Image)

	for _, l := range job.Commands {
		fmt.Println(l)
	}

	info.Language = "python361"
	job, err = ParseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\nim:", job.Image)

	for _, l := range job.Commands {
		fmt.Println(l)
	}
}
