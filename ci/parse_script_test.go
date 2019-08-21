package ci

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"
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
	getURL := "https://github.com/dat520-s2020/assignments.git"
	testURL := "https://github.com/dat520-s2020/tests.git"
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
	j, err := ParseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("TEST_TMPL") != "" {
		for _, cmd := range j.Commands {
			fmt.Println(cmd)
		}
	}

	info.Language = "python361"
	_, err = ParseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}

	info.Language = "java8"
	_, err = ParseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}
}
