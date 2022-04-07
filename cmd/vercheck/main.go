package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/autograde/quickfeed/kit/sh"
)

const (
	pbgo   = "ag/ag.pb.go"
	grpcpb = "ag/ag_grpc.pb.go"
)

func main() {
	protoc := regexp.MustCompile(`^\/\/.*(protoc)\s+v(.*)$`)
	genGo := regexp.MustCompile(`^\/\/.*(protoc-gen-go)\s+v(.*)$`)
	genGoGrpc := regexp.MustCompile(`^\/\/.*(protoc-gen-go-grpc)\s+v(.*)$`)

	needUpdate := false
	for re, file := range map[*regexp.Regexp]string{
		protoc:    pbgo,
		genGo:     pbgo,
		genGoGrpc: grpcpb,
	} {
		tool, codeVer := scan(file, re)
		toolVer := toolVersion(tool)
		needUpdate = needUpdate || checkVersions(tool, toolVer, codeVer)
	}
	if needUpdate {
		os.Exit(1)
	}
}

// checkVersions returns true if the installed tool must be updated.
func checkVersions(tool, toolVer, codeVer string) bool {
	if toolVer != codeVer && sort.StringsAreSorted([]string{toolVer, codeVer}) {
		fmt.Printf("Installed %s version %v is older than generated code version %v\n", tool, toolVer, codeVer)
		return true
	}
	return false
}

// toolVersion returns the given tool's version.
func toolVersion(tool string) string {
	s, err := sh.Output(tool + " --version")
	check(err)
	s = strings.TrimSpace(s)
	switch s {
	case "Missing value for flag: --version":
		fallthrough
	case `unknown argument "--version"`:
		fallthrough
	case `flag provided but not defined: -version`:
		log.Printf("Your installed %s version is too old. Please update to the latest version.", tool)
		return "0.0.0"
	}
	i := strings.LastIndex(s, " ")
	s = s[i+1:]
	if strings.HasPrefix(s, "v") {
		// annoyingly some tools use v and others don't
		return s[1:]
	}
	return s
}

// scan returns the tool (matching the regex) and version used to generated the given file.
func scan(file string, re *regexp.Regexp) (string, string) {
	f, err := os.Open(file)
	check(err)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			s := re.ReplaceAllString(line, "$1:$2")
			i := strings.Index(s, ":")
			return s[:i], s[i+1:] // tool and version
		}
	}
	check(scanner.Err())
	return "", ""
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
