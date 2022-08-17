package env

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/quickfeed/quickfeed/kit/sh"
)

const dotEnvPath = ".env"

var quickfeedRoot string

func init() {
	quickfeedRoot = os.Getenv("QUICKFEED")
	if quickfeedRoot == "" {
		out, err := sh.Output("go list -m -f {{.Dir}}")
		if err != nil {
			log.Fatalf("Failed to set QUICKFEED variable: %v", err)
		}
		quickfeedRoot = strings.TrimSpace(out)
		os.Setenv("QUICKFEED", quickfeedRoot)
	}
}

// Load loads environment variables from the given file, or from $QUICKFEED/.env.
// The variable's values are expanded with existing variables from the environment.
// It will not override a variable that already exists in the environment.
func Load(filename string) error {
	file, err := open(filename, false)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if ignore(line) {
			continue
		}
		key, val, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		k := strings.TrimSpace(key)
		if os.Getenv(k) != "" {
			// Ignore .env entries already set in the environment.
			continue
		}
		val = os.ExpandEnv(strings.Trim(strings.TrimSpace(val), `"`))
		os.Setenv(k, val)
	}

	return scanner.Err()
}

func open(filename string, write bool) (*os.File, error) {
	if filename == "" {
		filename = filepath.Join(quickfeedRoot, dotEnvPath)
		log.Printf("Loading environment variables from %s", filename)
	}
	if write {
		return os.OpenFile(filename, os.O_APPEND, 0644)
	}
	return os.Open(filename)
}

func Save(filename string, kv ...string) {
	if len(kv)%2 != 0 {
		log.Fatalf("Save: invalid number of arguments: %d", len(kv))
	}

	content, err := os.ReadFile(dotEnvPath)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", dotEnvPath, err)
	}
	for i := 0; i < len(kv); i += 2 {
		// Regexp to match key=value lines.
		re := regexp.MustCompile(fmt.Sprintf("(?m)^%s=.*$", kv[i]))
		key := kv[i]
		val := kv[i+1]
		if re.Match(content) {
			fmt.Println("Match")
		}
		if strings.Contains(string(content), key) {
			fmt.Println("Replacing", key, "with", val)
			// replace existing previous value
			fmt.Println(re.Match(content))
			content = re.ReplaceAll(content, []byte(key+"="+val+"\n"))
			continue
		}
		// append new value
		content = append(content, []byte("\n"+key+"="+val)...)
	}
	if err := os.Rename(dotEnvPath, dotEnvPath+".bak"); err != nil {
		log.Fatalf("Failed to rename %s: %v", dotEnvPath, err)
	}
	if err := os.WriteFile(dotEnvPath, content, 0644); err != nil {
		log.Fatalf("Failed to write %s: %v", dotEnvPath, err)
	}
}

func ignore(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}
