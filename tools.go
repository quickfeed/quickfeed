//go:build tools
// +build tools

package main

import (
	_ "github.com/alta/protopatch/cmd/protoc-gen-go-patch"
	_ "github.com/golang/protobuf/protoc-gen-go"
)
