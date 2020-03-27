package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var secretGo = `package {{ .Package }}

import (
  "io/ioutil"
  "log"

  "github.com/autograde/aguis/kit/score"
)

func init() {
  score.GlobalSecret = secret()
  log.SetOutput(ioutil.Discard)
}

func secret() string {
  return "{{ .RandomSecret }}"
}
`

var secretFile = "secret_ag_test.go"

func main() {
	var (
		agpath  = flag.String("path", ".", "directory to traverse for creating "+secretFile)
		secret  = flag.String("secret", "", "secret string to insert in the "+secretFile)
		verbose = flag.Bool("verbose", false, "be verbose")
	)
	flag.Parse()

	type SecretInfo struct {
		RandomSecret string
		Package      string
	}
	err := filepath.Walk(*agpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("failure accessing a path %q: %v\n", path, err)
			return err
		}
		if strings.HasPrefix(path, ".git") {
			// don't traverse .git
			return nil
		}
		if info.IsDir() && info.Name() != "." {
			// traverse subfolders of . (but not . itself)
			t, err := template.New("secret").Parse(secretGo)
			if err != nil {
				return err
			}
			buffer := new(bytes.Buffer)
			secretInfo := SecretInfo{RandomSecret: *secret, Package: info.Name()}
			if err := t.Execute(buffer, secretInfo); err != nil {
				return err
			}

			// format secret_ag_test.go to be canonical Go format
			formattedBuf, err := format.Source(buffer.Bytes())
			if err != nil {
				fmt.Printf("error formating source code for %s: %+v", secretFile, err)
				return err
			}

			err = ioutil.WriteFile(filepath.Join(path, secretFile), formattedBuf, 0644)
			if err != nil {
				return err
			}
			if *verbose {
				fmt.Printf("created: %s\n", filepath.Join(path, secretFile))
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", *agpath, err)
		return
	}
}
