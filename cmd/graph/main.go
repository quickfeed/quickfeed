package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	const pathToQuickfeedRoot = "../../"
	var wantedFiles []string // stores the paths of all wanted files
	if err := extractFilesInDirectory(pathToQuickfeedRoot, &wantedFiles); err != nil {
		fmt.Println(err)
		return
	}

	type symbol struct {
		name     string
		kind     string
		position string
	}

	cmd := exec.Command("gopls", "symbols", wantedFiles[0])
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}

	var symbols []symbol
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		args := strings.Split(line, " ")
		// Skip, if the line does not contain the expected number of arguments
		if len(args) == 3 { // There a cases of arrays with a single empty string entry
			// Append the symbol to the slice
			symbols = append(symbols, symbol{
				name:     args[0],
				kind:     args[1],
				position: args[2],
			})
		}
	}

	for _, s := range symbols {
		fmt.Printf("Name: %s\nKind: %s\nLocation: %s\n\n", s.name, s.kind, s.position)
	}
}

// extractFilesInDirectory extracts all files with the extensions .go, .ts, and .tsx in the given directory.
// function is recursive and will traverse all subdirectories.
func extractFilesInDirectory(dirPath string, wantedFiles *[]string) error {
	entities, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("Error reading directory: %v", err)
	}
	for _, entity := range entities {
		entityPath := filepath.Join(dirPath, entity.Name())
		if entity.IsDir() {
			extractFilesInDirectory(entityPath, wantedFiles)
			continue
		} else {
			if validateFile(entity) {
				*wantedFiles = append(*wantedFiles, entityPath)
			}
		}
	}
	return nil
}

// checks if the file is of type .go, .ts, or .tsx
func validateFile(file os.DirEntry) bool {
	fileName := file.Name()
	// return early if file does not contain a file extension
	if !strings.Contains(fileName, ".") {
		return false
	}
	// using bool map to easily check if file is of wanted extension
	// there probably a simpler way to define this map
	wantedExtensions := map[string]bool{"go": true, "ts": true, "tsx": true}
	return wantedExtensions[getFileExtension(fileName)]
}

// splits the file name and returns the file extension
func getFileExtension(fileName string) string {
	return strings.Split(fileName, ".")[1]
}
