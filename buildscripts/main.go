package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/autograde/aguis/models"
)

type NameStore struct {
	Name string
}

func main() {
	t, err := template.ParseFiles("go.tml")
	if err != nil {
		fmt.Println(err)
	}

	buffer := bytes.NewBufferString("")

	t.Execute(buffer, &models.AssignmentCIInfo{
		AccessToken:    "123456",
		AssignmentName: "GO",
		GetURL:         "https://github.com/somerepo",
		TestURL:        "https://github.com/somerepo-test",
	})

	lines := strings.Split(buffer.String(), "\n")
	restData, _, image := ExtractDockerImageInformation(lines)

	fmt.Println("Image:", *image)
	fmt.Println("Data:", restData)
}

func ExtractDockerImageInformation(lines []string) (data []string, container *string, image *string) {
	if len(lines) > 0 && strings.Index(lines[0], "#!") == 0 {
		firstLine := lines[0]
		rest := lines[1:]
		parts := strings.Split(firstLine, "/")
		if len(parts) > 2 && parts[1] == "docker" {
			return rest, &parts[1], &parts[2]
		}

	}
	return lines, nil, nil
}
