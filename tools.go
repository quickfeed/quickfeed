// +build tools

package main

import (
	_ "github.com/gogo/protobuf/gogoproto"
	_ "github.com/gogo/protobuf/protoc-gen-gofast"
	_ "github.com/gogo/protobuf/protoc-gen-gogofast"
	_ "github.com/gogo/protobuf/protoc-gen-gogofaster"
	_ "github.com/golang/protobuf/protoc-gen-go"
)
