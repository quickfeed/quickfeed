package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
	This program is highly dependent on the gopls library.
*/

func main() {
	fmt.Println("Creating visual graph of Quickfeed, this may take a while...")

	const pathToQuickfeedRoot = "../../"
	var wantedFiles []string // stores the paths of all wanted files
	fmt.Println("Extracting files in directory...")
	if err := extractFilesInDirectory(pathToQuickfeedRoot, &wantedFiles); err != nil {
		fmt.Println(err)
		return
	}

	symbolsMap := make(map[string][]symbol) // Maps symbols to their respective files
	fmt.Println("Creating symbol map...")
	if err := createSymbolMap(wantedFiles[0], &symbolsMap); err != nil {
		fmt.Println(err)
		return
	}

	for _, symbols := range symbolsMap {
		for _, symbol := range symbols {
			fmt.Printf("%s %s %s\n", symbol.name, symbol.kind, symbol.position)
			for _, reference := range symbol.refs {
				fmt.Printf("\t -- %s\n", reference)
			}
		}
	}
}

type position struct {
	line      string
	charRange string
}

func (p position) getPos() string {
	return fmt.Sprintf("%s:%s", p.line, p.charRange)
}

type symbol struct {
	name     string
	kind     string
	position position
	refs     []string
}

// creates a map of all symbols in the wanted files
// the key is the file path and the value is a slice of symbols in the file
func createSymbolMap(filePath string, symbolsMap *map[string][]symbol) error {
	/*for _, filePath := range wantedFiles {*/
	if output, err := exec.Command("gopls", "symbols", filePath).Output(); err != nil {
		return err
	} else {
		fmt.Printf("Operating on file:  %s, File path: %s\n", getLastEntry(filePath, "/"), filePath)
		fmt.Println("Extracting symbols from file")
		(*symbolsMap)[filePath] = extractSymbols(string(output))

		fmt.Println("Finding all references in file")
		if err := findRefs(symbolsMap, filePath); err != nil {
			return err
		}
	}
	/*}*/
	return nil
}

// returns the last entry of a string array with a given delimiter
func getLastEntry(str string, delimiter string) string {
	split := strings.Split(str, delimiter)
	return split[len(split)-1]
}

// Loops through all symbols and finds references for each symbol
func findRefs(symbolsMap *map[string][]symbol, filePath string) error {
	for i := range (*symbolsMap)[filePath] {
		fmt.Printf("Executing references command for symbol: %s\n", (*symbolsMap)[filePath][i].name)
		pathToSymbol := fmt.Sprintf("%s:%s", filePath, (*symbolsMap)[filePath][i].position.getPos())
		fmt.Printf("Path to symbol: %s\n", pathToSymbol)
		if output, err := exec.Command("gopls", "references", pathToSymbol).Output(); err != nil {
			return err
		} else {
			fmt.Println("Extracting symbols from file (references)")

			for _, line := range strings.Split(string(output), "\n") {
				if line == "" {
					continue
				}
				(*symbolsMap)[filePath][i].refs = append((*symbolsMap)[filePath][i].refs, line)
			}
		}
	}
	return nil
}

// parses the output of the gopls symbols command and extracts the name, kind, and position of each symbol
func extractSymbols(output string) []symbol {
	// Gets the line and character range position of the symbol
	getPosition := func(_position string) position {
		args := strings.Split(_position, "-")
		sLineP := strings.Split(args[0], ":")[0] // starting line position
		sCharP := strings.Split(args[0], ":")[1] // starting character position
		eCharP := strings.Split(args[1], ":")[0] // ending character position
		return position{line: sLineP, charRange: fmt.Sprintf("%s-%s", sCharP, eCharP)}
	}

	var symbols []symbol
	for _, line := range strings.Split(output, "\n") {
		args := strings.Split(line, " ")
		// Skip, if the line does not contain the expected number of arguments
		if len(args) == 3 { // There a cases of arrays with a single empty string entry
			symbols = append(symbols, symbol{
				name:     args[0],
				kind:     args[1],
				position: getPosition(args[2]),
			})
		}
	}
	return symbols
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
