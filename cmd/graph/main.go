package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	rootFolderName = "quickfeed"
	function       = "Function"
	method         = "Method"
)

func main() {
	old := flag.Bool("old", false, "create map with old data")
	flag.Parse()
	const rePath = "../../../" // Relative path to Quickfeed folder
	cacheFilePath := "map.json"
	projectMap := make(fMap)
	if *old {
		if err := projectMap.getContentFromCache(cacheFilePath); err != nil {
			fmt.Printf("Error getting content from cache: %v\n", err)
		}
	} else {
		fmt.Println("Creating visual graph of Quickfeed, this will take a while...")
		/*
			TODO: Make the error handling return the gopls log instead of the error status message..
		*/
		if err := projectMap.createMap(rePath, cacheFilePath); err != nil {
			fmt.Printf("Error creating map: %v\n", err)
			return
		}
	}
	/*
		Following can be written with any graphing library
		Currently, the graph is visualized with graphviz
	*/
	// Initialize the graph file, delete if it already exists
	graphFilePath := "qf-graph.dot"
	if _, err := os.Stat(graphFilePath); !os.IsNotExist(err) {
		if err := os.Remove(graphFilePath); err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (fMap *fMap) getContentFromCache(cacheFilePath string) error {
	if bytes, err := os.ReadFile(cacheFilePath); err != nil {
		return err
	} else {
		if err := json.Unmarshal(bytes, &fMap); err != nil {
			return err
		}
	}
	return nil
}

func (fMap *fMap) createMap(rePath string, cacheFilePath string) error {
	if err := getContent(fMap, rePath, "", nil); err != nil {
		return fmt.Errorf("Error when getting content: %v", err)
	}
	if err := clean(fMap); err != nil {
		return fmt.Errorf("Error when cleaning.. err: %s", err)
	}
	if err := populate(fMap); err != nil {
		return fmt.Errorf("Error populating: %v", err)
	}

	/* TODO: Fix marshalling of the map */
	projectMapJSON, err := json.MarshalIndent(*fMap, "", "\t")
	if err != nil {
		return fmt.Errorf("Error marshalling projectMap: %v", err)
	}
	if _, err := os.Stat(cacheFilePath); !os.IsNotExist(err) {
		if err := os.Remove(cacheFilePath); err != nil {
			return err
		}
	}
	if err := os.WriteFile(cacheFilePath, projectMapJSON, 0o644); err != nil {
		return err
	}

	return nil
}

// remove entries with zero files and subfolders
func clean(fMap *fMap) error {
	var keysToDelete []string
	for key, folder := range *fMap {
		if len(folder.SubFolders) == 0 {
			if len(folder.Files) == 0 {
				keysToDelete = append(keysToDelete, key)
			}
		} else {
			clean(&folder.SubFolders)
		}
	}
	for _, key := range keysToDelete {
		delete(*fMap, key)
	}
	return nil
}

// Populate gathers all the symbols and references in the project and structures them relative to where the reference is found
// If the reference is outside the folder, its stored in the ref slice in the folder struct
// If the reference is found in a different file but same parent folder, its stored in the ref slice in the file struct
// If the reference is found in the same file, its stored in the ref slice in the symbol struct
func populate(fMap *fMap) error {
	for _, folder := range *fMap {
		for i, file := range folder.Files {
			fmt.Printf("Operating on file:  %s, File path: %s\n", file.Name, file.Path)
			if err := folder.setSymbols(file.Path, i); err != nil {
				return err
			}
			fmt.Println("Finding all references to symbols in file")
			if err := folder.findRefs(file.Path, i, *fMap); err != nil {
				return err
			}
		}
		populate(&folder.SubFolders)
	}
	return nil
}

// runs the gopls command with the given arguments
func runGopls(args ...string) ([]byte, error) {
	_args := []string{"-vv", "-rpc.trace"}
	return exec.Command("gopls", append(_args, args...)...).Output()
}

// gets all symbols in a file if the file is not already in the symbols map
func (f folder) setSymbols(filePath string, fileIndex int) error {
	if len(f.Files[fileIndex].Symbols) == 0 {
		output, err := runGopls("symbols", filePath)
		if err != nil {
			return err
		}
		fmt.Printf("Extracting symbols from file, path: %s\n", filePath)
		f.Files[fileIndex].Symbols = extractSymbols(string(output))
	}
	return nil
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
		return position{
			Line:      sLineP,
			CharRange: fmt.Sprintf("%s-%s", sCharP, eCharP),
		}
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
				Name:     name,
				Kind:     kind,
				Position: createPosition(args[2]),
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
	symbols := folder.Files[fileIndex].Symbols
	for i := range symbols {
		pathToSymbol := fmt.Sprintf("%s:%s", filePath, symbols[i].Position.getPos())
		if output, err := runGopls("references", pathToSymbol); err != nil {
			return err
		} else {
			var refs []ref
			if err := findParentsForRefs(&refs, createRefInfo(filePath, symbols[i].Name), parseRefs(string(output)), fMap); err != nil {
				return err
			}
			if err := folder.assignRefsToMap(fileIndex, i, refs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (folder folder) assignRefsToMap(fileIndex int, symbolIndex int, refs []ref) error {
	for _, ref := range refs {
		if ref.Source.FolderName == ref.Info.FolderName {
			if ref.Source.FileName == ref.Info.FileName {
				folder.Files[fileIndex].Symbols[symbolIndex].Refs = append(folder.Files[fileIndex].Symbols[symbolIndex].Refs, ref)
			} else {
				folder.Files[fileIndex].Refs = append(folder.Files[fileIndex].Refs, ref)
			}
		} else {
			folder.Refs = append(folder.Refs, ref)
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

func createRefInfo(filePath string, symbolName string) refInfo {
	args := strings.Split(filePath, ":")
	filePath = args[0]
	fileName := getLastEntry(filePath, "/", 0)
	folderName := getLastEntry(filePath, "/", 1)
	return refInfo{Path: filePath, FolderName: folderName, FileName: fileName, MethodName: symbolName}
}

func getKeys(filePath string) ([]string, error) {
	args := strings.Split(filePath, "/")
	for i := len(args) - 1; i >= 0; i-- {
		if args[i] == rootFolderName {
			return args[i : len(args)-1], nil
		}
	}
	return nil, fmt.Errorf("Could not find quickfeed directory in path: %s", filePath)
}

func (fMap fMap) getFolderAndFileIndex(filePath string, fileName string) (folder, int, error) {
	keys, err := getKeys(filePath)
	if err != nil {
		return folder{}, 0, err
	}
	var folder folder
	for i, key := range keys {
		folder = fMap[key]
		// if the folder has subfolders, update the folder to the subfolder
		// only if the current key is not the last key
		if folder.SubFolders != nil && i < len(keys)-1 {
			fMap = folder.SubFolders
		}
	}
	for i, file := range folder.Files {
		if file.Name == fileName {
			return folder, i, nil
		}
	}
	return folder, 0, fmt.Errorf("Could not find file: %s in folder: %s", fileName, folder.FolderPath)
}

// finds the parent method for each reference
func findParentsForRefs(parent_refs *[]ref, source_refInfo refInfo, refs []string, fMap fMap) error {
	for _, _ref := range refs {
		args := strings.Split(_ref, ":")
		filePath := args[0]
		fileName := getLastEntry(filePath, "/", 0)
		linePos := args[1]
		// get symbols in the reference file
		// if the file is not already in the symbol map
		folder, fileIndex, err := fMap.getFolderAndFileIndex(filePath, fileName)
		if err != nil {
			return err
		}
		if err := folder.setSymbols(filePath, fileIndex); err != nil {
			return err
		}
		// closest method above symbol, initial value is a symbol with line 0
		refParent := symbol{Position: position{Line: "0"}}
		if err := getRelatedMethod(folder.Files[fileIndex].Symbols, &refParent, linePos); err != nil {
			return err
		}
		refInfo := createRefInfo(filePath, refParent.Name)
		*parent_refs = append(*parent_refs, ref{Source: source_refInfo, Info: refInfo})

	}
	return nil
}

// returns entry relative to last, of a string array with a given delimiter, i determines how many entries from the end
func getLastEntry(str string, delimiter string, i int) string {
	split := strings.Split(str, delimiter)
	return split[len(split)-1-i]
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
		if p_symbol.Kind != function && p_symbol.Kind != method {
			continue
		}
		newMethodLinePos, err := parseStringToInt(p_symbol.Position.Line)
		if err != nil {
			return err
		}
		currentMethodLinePos, err := parseStringToInt(refParent.Position.Line)
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
	FolderPath string `json:"folderPath"`
	Refs       []ref  `json:"refs"`
	Files      []file `json:"files"`
	SubFolders fMap   `json:"subFolders"`
}

type file struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Refs    []ref    `json:"refs"`
	Symbols []symbol `json:"symbols"`
}

type symbol struct {
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Position position `json:"position"`
	Refs     []ref    `json:"refs"`
}

// source is there since the symbol is a reference to a symbol in another file
// Will result in duplicate data, but it's needed to keep track of the source
type ref struct {
	Source refInfo `json:"source"`
	Info   refInfo `json:"info"`
}

type refInfo struct {
	Path       string `json:"path"`
	FolderName string `json:"folderName"`
	FileName   string `json:"fileName"`
	MethodName string `json:"methodName"`
}

type position struct {
	Line      string `json:"line"`
	CharRange string `json:"charRange"`
}

func (p position) getPos() string {
	return fmt.Sprintf("%s:%s", p.Line, p.CharRange)
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
			(*childMap)[name] = folder{FolderPath: subDirPath}
			if entry, ok := (*childMap)[name]; ok {
				entry.SubFolders = make(fMap)
				if err := getContent(&entry.SubFolders, subDirPath, name, childMap); err != nil {
					return err
				}
				/*
					Get the updated entry from childMap and combine with parentMap updates.
					The files are added concurrently to the parentMap which is pointing to the same map as the childMap.
					This is needed to get the updated entry (after running getContent recursively) from the common method which is initialized in the main method.

					There is probably a better way to do this, but I will leave it as is for now.
				*/
				if second_entry, ok := (*childMap)[name]; ok {
					second_entry.SubFolders = entry.SubFolders
					(*childMap)[name] = second_entry
				}
			}
		} else {
			subDirPath := filepath.Join(dirPath, name)
			if parentDirName == "" {
				return fmt.Errorf("Parent directory name can't be empty..")
			}
			if folder, ok := (*parentMap)[parentDirName]; ok {
				folder.Files = append(folder.Files, file{Name: name, Path: subDirPath})
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
		// includeDirs := map[string]bool{"assignments": true, rootFolderName: true}
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
