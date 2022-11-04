//go:build tools

package main

import (
	_ "github.com/alta/protopatch/cmd/protoc-gen-go-patch"
	_ "github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
