OS					:= $(shell echo $(shell uname -s) | tr A-Z a-z)
ARCH				:= $(shell uname -m)
tmpdir				:= tmp
proto-path			:= public/proto
grpcweb-ver			:= 1.0.4
protoc-grpcweb		:= protoc-gen-grpc-web
protoc-grpcweb-long	:= $(protoc-grpcweb)-$(grpcweb-ver)-$(OS)-$(ARCH)
grpcweb-url			:= https://github.com/grpc/grpc-web/releases/download/$(grpcweb-ver)/$(protoc-grpcweb-long)
grpcweb-path		:= /usr/local/bin/$(protoc-grpcweb)

# necessary when target is not tied to a file
.PHONY: dep install ui proto devtools grpcweb

dep:
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/gogo/protobuf/proto
	go get -u github.com/gogo/protobuf/gogoproto
	go get -u github.com/gogo/protobuf/protoc-gen-gofast
	go get -u github.com/gogo/protobuf/protoc-gen-gogofast
	go get -u github.com/gogo/protobuf/protoc-gen-gogofaster

install:
	@echo go install
	@go install

ui:
	@echo Running webpack
	@cd public; npm install; webpack

proto:
	@echo Compiling Autograders proto definitions
	@cd ag; protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf --gogofast_out=plugins=grpc,\
	Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:. \
	--js_out=import_style=commonjs:../$(proto-path)/ \
	--grpc-web_out=import_style=typescript,mode=grpcweb:../$(proto-path)/ ag.proto
	@sed -i '' '/gogo/d' $(proto-path)/ag_pb.js $(proto-path)/AgServiceClientPb.ts $(proto-path)/ag_pb.d.ts
	@tsc $(proto-path)/AgServiceClientPb.ts

devtools: grpcweb

grpcweb:
	@echo Fetch and install grpcweb protoc plugin (requires sudo access)
	@mkdir -p $(tmpdir)
	@cd $(tmpdir); curl -LOs $(grpcweb-url)
	@sudo mv $(tmpdir)/$(protoc-grpcweb-long) $(grpcweb-path)
	@chmod +x $(grpcweb-path)
	@rm -rf $(tmpdir)
