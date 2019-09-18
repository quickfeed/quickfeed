package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

var secretGo = `package {{ .Package }}

import (
  "io/ioutil"
  "log"

  "github.com/autograde/kit/score"
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
		agpath = flag.String("path", ".", "directory to traverse for creating "+secretFile)
		secret = flag.String("secret", "", "secret string to insert in the "+secretFile)
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
		if info.IsDir() && info.Name() != "." {
			t, err := template.New("secret").Parse(secretGo)
			if err != nil {
				return err
			}
			buffer := new(bytes.Buffer)
			secretInfo := SecretInfo{RandomSecret: *secret, Package: info.Name()}
			if err := t.Execute(buffer, secretInfo); err != nil {
				return err
			}

			err = ioutil.WriteFile(filepath.Join(path, secretFile), buffer.Bytes(), 0644)
			if err != nil {
				return err
			}
			fmt.Printf("created: %s\n", filepath.Join(path, secretFile))
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", *agpath, err)
		return
	}
}
