package ci

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"strings"
	"testing"
)

func TestParseScript(t *testing.T) {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		t.Fatal(err)
	}
	randomString := fmt.Sprintf("%x", sha1.Sum(randomness))
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
		RandomSecret:       randomString,
	}
	job, err := ParseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range job.Commands {
		fmt.Println(l)
	}

	info.Language = "python361"
	job, err = ParseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}
}
