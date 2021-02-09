package ci

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"
	"testing"
)

func TestParseScript(t *testing.T) {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		t.Fatal(err)
	}
	randomString := fmt.Sprintf("%x", sha1.Sum(randomness))
	info := &AssignmentInfo{
		AssignmentName:     "lab2",
		Script:             "go.sh",
		CreatorAccessToken: "secret",
		GetURL:             getURL,
		TestURL:            testURL,
		RandomSecret:       randomString,
	}
	j, err := parseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("TEST_TMPL") != "" {
		for _, cmd := range j.Commands {
			fmt.Println(cmd)
		}
	}
	if os.Getenv("TEST_IMAGE") != "" {
		fmt.Println(j.Image)
	}

	info.Script = "python361.sh"
	_, err = parseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}

	info.Script = "java8.sh"
	_, err = parseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}

	info.Script = "python-dat550.sh"
	j, err = parseScriptTemplate("scripts", info)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("TEST_TMPL") != "" {
		for _, cmd := range j.Commands {
			fmt.Println(cmd)
		}
	}
	if os.Getenv("TEST_IMAGE") != "" {
		fmt.Println(j.Image)
	}
}
