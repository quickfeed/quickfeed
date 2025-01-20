package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	const rePath = "../../../" // Relative path to Quickfeed folder

	projectMap := make(fMap)
	if err := getContent(&projectMap, rePath, "", nil); err != nil {
		fmt.Printf("Error getting content: %v\n", err)
		return
	}
	if err := clean(projectMap); err != nil {
		fmt.Printf("Error when cleaning.. err: %s", err)
		return
	}
	if err := populate(projectMap); err != nil {
		fmt.Printf("Error populating: %v\n", err)
		return
	}

	fmt.Println(projectMap)
}

// remove entries with zero files and subfolders
func clean(fMap fMap) error {
	var keysToDelete []string
	for key := range fMap {
		if len(fMap[key].subFolders) == 0 {
			if len(fMap[key].files) == 0 {
				keysToDelete = append(keysToDelete, key)
			}
		} else {
			clean(fMap[key].subFolders)
		}
	}
	for _, key := range keysToDelete {
		delete(fMap, key)
	}
	return nil
}

const (
	constant = "Constant"
	variable = "Variable"
	function = "Function"
	method   = "Method"
)

// Populate gathers all the symbols and references in the project and structures them relative to where the reference is found
// If the reference is outside the folder, its stored in the ref slice in the folder struct
// If the reference is found in a different file but same parent folder, its stored in the ref slice in the file struct
// If the reference is found in the same file, its stored in the ref slice in the symbol struct
func populate(fMap fMap) error {
	for key := range fMap {
		for i, file := range fMap[key].files {
			fmt.Printf("Operating on file:  %s, File path: %s\n", file.name, file.path)
			symbols, err := fMap.getSymbols(file.path)
			if err != nil {
				return err
			}
			fMap[key].files[i].symbols = symbols
			fmt.Println("Finding all references in file")
			if err := fMap[key].findRefs(file.path, i, fMap); err != nil {
				return err
			}
		}
		populate(fMap[key].subFolders)
	}
	return nil
}

// runs the gopls command with the given arguments
func runGopls(args ...string) ([]byte, error) {
	_args := []string{"-vv", "-rpc.trace"}
	return exec.Command("gopls", append(_args, args...)...).Output()
}

// gets all symbols in a file if the file is not already in the symbols map
func (fMap *fMap) getSymbols(filePath string) ([]symbol, error) {
	output, err := runGopls("symbols", filePath)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Extracting symbols from file, path: %s\n", filePath)
	return extractSymbols(string(output)), nil
}

// parses the output of the gopls symbols command and extracts the name, kind, and position of each symbol
func extractSymbols(output string) []symbol {
	// Gets the line and character range position of the symbol
	createPosition := func(_position string) position {
		args := strings.Split(_position, "-")
		args2 := strings.Split(args[0], ":")
		sLineP := args2[0]                       // starting line position
		sCharP := args2[1]                       // starting character position
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

func parseStringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// Loops through all symbols and finds references for each symbol
// Cases where the symbol is a function, the call hierarchy is also extracted
func (folder folder) findRefs(filePath string, fileIndex int, fMap fMap) error {
	for i := range folder.files[fileIndex].symbols {
		symbol := folder.files[fileIndex].symbols[i]
		fmt.Printf("Executing references command for symbol: %s\n", symbol.name)
		pathToSymbol := fmt.Sprintf("%s:%s", filePath, symbol.position.getPos())
		fmt.Printf("Path to symbol: %s\n", pathToSymbol)
		if output, err := runGopls("references", pathToSymbol); err != nil {
			return err
		} else {
			fmt.Println("Finding parent methods of references")
			if err := folder.findParentsForRefs(parseRefs(string(output)), i, fileIndex, fMap); err != nil {
				return err
			}
		}
	}
	return nil
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

// finds the parent method for each reference
func (folder *folder) findParentsForRefs(refs []string, symbolIndex int, fileIndex int, fMap fMap) error {
	for _, _ref := range refs {
		args := strings.Split(_ref, ":")
		refFile := args[0]
		refLinePos := args[1]
		// get symbols in the reference file
		// if the file is not already in the symbol map
		/*
			TODO: extract the directories and key into the map and check if the symbols are already in the map, if not, get the symbols.
			This can improve performance
		*/
		if symbols, err := fMap.getSymbols(refFile); err != nil {
			return err
		} else {
			// closest method above symbol, initial value is a symbol with line 0
			refParent := symbol{position: position{line: "0"}}
			if err := getRelatedMethod(symbols, &refParent, refLinePos); err != nil {
				return err
			}
			ref := ref{symbolPath: _ref, parent: refParent}
			folder.files[fileIndex].symbols[symbolIndex].refs = append(folder.files[fileIndex].symbols[symbolIndex].refs, ref)
		}
	}
	return nil
}

// getRelatedMethod finds the closest method above the reference
func getRelatedMethod(symbols []symbol, refParent *symbol, refLinePos string) error {
	_refLinePos, err := parseStringToInt(refLinePos)
	if err != nil {
		return err
	}
	// loop through potential parent symbols
	for _, p_symbol := range symbols {
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

type fMap map[string]folder

type folder struct {
	folderPath string
	refs       []ref
	files      []file
	subFolders fMap
}

type file struct {
	name    string
	path    string
	refs    []ref
	symbols []symbol
}

type folderRef struct {
	filePath string
	symbols  []symbol
}

type symbol struct {
	name     string
	kind     string
	position position
	refs     []ref
}

type ref struct {
	symbolPath string
	parent     symbol
}

type position struct {
	line      string
	charRange string
}

func (p position) getPos() string {
	return fmt.Sprintf("%s:%s", p.line, p.charRange)
}

// getContent recursively reads the content of a directory and its subdirectories
// Should be refactored... but not sure how atm
func getContent(childMap *fMap, dirPath string, parentDirName string, parentMap *fMap) error {
	entities, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("Error reading directory: %v", err)
	}
	for _, entity := range entities {
		if !isValid(entity) {
			continue
		}
		name := entity.Name()
		if entity.IsDir() {
			subDirPath := filepath.Join(dirPath, name)
			(*childMap)[name] = folder{folderPath: subDirPath}
			if entry, ok := (*childMap)[name]; ok {
				entry.subFolders = make(fMap)
				if err := getContent(&entry.subFolders, subDirPath, name, childMap); err != nil {
					return err
				}
				/*
					Get the updated entry from childMap and combine with parentMap updates.
					The files are added concurrently to the parentMap which is pointing to the same map as the childMap.
					This is needed to get the updated entry (after running getContent recursively) from the common method which is initialized in the main method.

					There is probably a better way to do this, but I will leave it as is for now.
				*/
				if second_entry, ok := (*childMap)[name]; ok {
					second_entry.subFolders = entry.subFolders
					(*childMap)[name] = second_entry
				}
			}
		} else {
			subDirPath := filepath.Join(dirPath, name)
			if parentDirName == "" {
				return fmt.Errorf("Parent directory name can't be empty..")
			}
			if folder, ok := (*parentMap)[parentDirName]; ok {
				folder.files = append(folder.files, file{name: name, path: subDirPath})
				(*parentMap)[parentDirName] = folder
			}
		}
	}
	return nil
}

// checks if the file is of type .go, .ts, or .tsx
func isValid(dirEntry os.DirEntry) bool {
	name := dirEntry.Name()
	// return early if directory entry does not contain a file extension
	if !strings.Contains(name, ".") {
		// limit to only include directories with the following names
		// includeDirs := map[string]bool{"assignments": true, "quickfeed": true}
		// return includeDirs[name]
		excludedDirs := map[string]bool{"node_modules": true}
		if excludedDirs[name] {
			return false
		} else {
			// Some files without a period is for some reason a directory
			// This will exclude those files
			// For example LICENSE does not contain a period and is a file, os thinks it's a directory
			return dirEntry.IsDir()
		}
	}
	// using bool map to easily check if file is of wanted extension
	// there probably a simpler way to define this map
	wantedExtensions := map[string]bool{"go": true, "ts": true, "tsx": true}
	return wantedExtensions[getFileExtension(name)]
}

// splits the file name and returns the file extension
func getFileExtension(fileName string) string {
	return strings.Split(fileName, ".")[1]
}
