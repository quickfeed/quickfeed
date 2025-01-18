package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

/*
	This program is highly dependent on the gopls library.
*/

func main() {
	fmt.Println("Creating visual graph of Quickfeed, this will take a while...")

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

	/*
		Following can be written with any graphing library

		Currently, the graph visualized with graphviz
	*/

	// Initialize the graph file, delete if it already exists
	graphFilePath := fmt.Sprintf("%sqf-graph.dot", pathToQuickfeedRoot)
	if _, err := os.Stat(graphFilePath); !os.IsNotExist(err) {
		if err := os.Remove(graphFilePath); err != nil {
			fmt.Println(err)
			return
		}
	}
}

const (
	constant = "Constant"
	variable = "Variable"
	function = "Function"
	method   = "Method"
)

type position struct {
	line      string
	charRange string
}

func (p position) getPos() string {
	return fmt.Sprintf("%s:%s", p.line, p.charRange)
}

type ref struct {
	symbolPath string
	parent     symbol
}

type symbol struct {
	name     string
	kind     string
	position position
	refs     []ref
}

// runs the gopls command with the given arguments
func runGopls(args ...string) ([]byte, error) {
	_args := []string{"-vv", "-rpc.trace"}
	return exec.Command("gopls", append(_args, args...)...).Output()
}

// creates a map of all symbols in the wanted files
// the key is the file path and the value is a slice of symbols in the file
func createSymbolMap(filePath string, symbolsMap *map[string][]symbol) error {
	/*for _, filePath := range wantedFiles {*/

	fmt.Printf("Operating on file:  %s, File path: %s\n", getLastEntry(filePath, "/"), filePath)
	if err := getSymbols(symbolsMap, filePath); err != nil {
		return err
	}

	fmt.Println("Finding all references in file")
	if err := findRefs(symbolsMap, filePath); err != nil {
		return err
	}
	/*}*/
	return nil
}

// gets all symbols in a file if the file is not already in the symbols map
func getSymbols(symbolsMap *map[string][]symbol, filePath string) error {
	get := func(filePath string) ([]symbol, error) {
		if output, err := runGopls("symbols", filePath); err != nil {
			return []symbol{}, err
		} else {
			fmt.Println("Extracting symbols from file")
			return extractSymbols(string(output)), nil
		}
	}
	if (*symbolsMap)[filePath] == nil {
		if symbols, err := get(filePath); err != nil {
			return err
		} else {
			(*symbolsMap)[filePath] = symbols
		}
	}
	return nil
}

// returns the last entry of a string array with a given delimiter
func getLastEntry(str string, delimiter string) string {
	split := strings.Split(str, delimiter)
	return split[len(split)-1]
}

func parseStringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// Loops through all symbols and finds references for each symbol
// Cases where the symbol is a function, the call hierarchy is also extracted
func findRefs(symbolsMap *map[string][]symbol, filePath string) error {
	for i := range (*symbolsMap)[filePath] {
		fmt.Printf("Executing references command for symbol: %s\n", (*symbolsMap)[filePath][i].name)
		pathToSymbol := fmt.Sprintf("%s:%s", filePath, (*symbolsMap)[filePath][i].position.getPos())
		fmt.Printf("Path to symbol: %s\n", pathToSymbol)
		if output, err := runGopls("references", pathToSymbol); err != nil {
			return err
		} else {
			// (*symbolsMap)[filePath][i].refs
			fmt.Println("Finding parent methods of references")
			if err := findParentsForRefs(symbolsMap, parseRefs(string(output)), i, filePath); err != nil {
				return err
			}
		}
	}
	return nil
}

// finds the parent method for each reference
func findParentsForRefs(symbolMap *map[string][]symbol, refs []string, i int, originalFile string) error {
	for _, ref := range refs {
		args := strings.Split(ref, ":")
		refFile := args[0]
		refLinePos := args[1]
		// get symbols in the reference file
		// if the file is not already in the symbol map
		if err := getSymbols(symbolMap, refFile); err != nil {
			return err
		}
		// closest method above symbol, initial value is a symbol with line 0
		refParent := symbol{position: position{line: "0"}}
		if err := getRelatedMethod(&refParent, symbolMap, refFile, refLinePos); err != nil {
			return err
		}
		(*symbolMap)[originalFile][i].refs = append((*symbolMap)[originalFile][i].refs, createRef(ref, refParent))
	}
	return nil
}

// getRelatedMethod finds the closest method above the reference
func getRelatedMethod(refParent *symbol, symbolsMap *map[string][]symbol, refFile string, refLinePos string) error {
	_refLinePos, err := parseStringToInt(refLinePos)
	if err != nil {
		return err
	}
	// loop through potential parent symbols
	for _, p_symbol := range (*symbolsMap)[refFile] {
		// skip if the symbol is not a function
		if p_symbol.kind != function && p_symbol.kind != method {
			continue
		}
		newMethodLinePos, err := parseStringToInt(p_symbol.position.line)
		if err != nil {
			return err
		}
		currentMethodLinePos, err := parseStringToInt(refParent.position.line)
		if err != nil {
			return err
		}
		newMethodIsFurtherDown := currentMethodLinePos < newMethodLinePos
		newMethodIsAboveRef := newMethodLinePos < _refLinePos
		if newMethodIsFurtherDown && newMethodIsAboveRef {
			*refParent = p_symbol
		}
	}
	return nil
}

func createRef(s string, p symbol) ref {
	return ref{symbolPath: s, parent: p}
}

func parseRefs(output string) []string {
	var refs []string
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		refs = append(refs, line)
	}
	return refs
}

// parses the output of the gopls symbols command and extracts the name, kind, and position of each symbol
func extractSymbols(output string) []symbol {
	// Gets the line and character range position of the symbol
	createPosition := func(_position string) position {
		args := strings.Split(_position, "-")
		sLineP := strings.Split(args[0], ":")[0] // starting line position
		sCharP := strings.Split(args[0], ":")[1] // starting character position
		eCharP := strings.Split(args[1], ":")[1] // ending character position
		return position{line: sLineP, charRange: fmt.Sprintf("%s-%s", sCharP, eCharP)}
	}
	var symbols []symbol
	for _, line := range strings.Split(output, "\n") {
		args := strings.Split(line, " ")
		// Skip, if the line does not contain the expected number of arguments
		if len(args) == 3 { // There a cases of arrays with a single empty string entry
			name := args[0]
			kind := args[1]
			if kind == method && strings.Contains(name, ".") {
				name = strings.Split(name, ".")[1]
			}
			symbols = append(symbols, symbol{
				name:     name,
				kind:     kind,
				position: createPosition(args[2]),
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
