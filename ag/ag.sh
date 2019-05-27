#!/bin/bash

# generates grpc files from ag.proto
# precompiles a generated ts for webpack compatibility
# removes gogo references from the js file


protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogofaster_out=\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types=\
plugins=grpc:. --js_out=import_style=commonjs:../public/proto/  --grpc-web_out=import_style=typescript,mode=grpcweb:../public/proto/ ag.proto

tsc ../public/proto/AgServiceClientPb.ts

sed -i '/gogo/d' ../public/proto/ag_pb.js
