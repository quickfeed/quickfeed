OS					:= $(shell echo $(shell uname -s) | tr A-Z a-z)
ARCH				:= $(shell uname -m)
tmpdir				:= tmp
proto-path			:= public/proto
grpcweb-ver			:= 1.0.4
protoc-grpcweb		:= protoc-gen-grpc-web
protoc-grpcweb-long	:= $(protoc-grpcweb)-$(grpcweb-ver)-$(OS)-$(ARCH)
grpcweb-url			:= https://github.com/grpc/grpc-web/releases/download/$(grpcweb-ver)/$(protoc-grpcweb-long)
grpcweb-path		:= /usr/local/bin/$(protoc-grpcweb)
sedi				:= $(shell sed --version >/dev/null 2>&1 && echo "sed -i --" || echo "sed -i ''")

# necessary when target is not tied to a file
.PHONY: dep install ui proto devtools grpcweb envoy-build envoy-run

dep:
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/gogo/protobuf/proto
	go get -u github.com/gogo/protobuf/gogoproto
	go get -u github.com/gogo/protobuf/protoc-gen-gofast
	go get -u github.com/gogo/protobuf/protoc-gen-gogofast
	go get -u github.com/gogo/protobuf/protoc-gen-gogofaster
	
# change back to 'go'
install:
	@echo go install
	@go1.13beta1 install

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
	$(sedi) '/gogo/d' $(proto-path)/ag_pb.js $(proto-path)/AgServiceClientPb.ts $(proto-path)/ag_pb.d.ts
	@tsc $(proto-path)/AgServiceClientPb.ts

devtools: grpcweb npmtools

grpcweb:
	@echo "Fetch and install grpcweb protoc plugin (requires sudo access)"
	@mkdir -p $(tmpdir)
	@cd $(tmpdir); curl -LOs $(grpcweb-url)
	@sudo mv $(tmpdir)/$(protoc-grpcweb-long) $(grpcweb-path)
	@chmod +x $(grpcweb-path)
	@rm -rf $(tmpdir)

npmtools:
	@echo "Install webpack and typescript compiler (requires sudo access)"
	@npm install -g --save typescript
	@npm install -g webpack
	@npm install -g webpack-cli
	@npm install -g tslint

envoy-build:
	@echo Building Autograder Envoy proxy
	@cd envoy; docker build -t ag_envoy -f ./envoy/envoy.Dockerfile .

envoy-run:
	@echo Starting Autograder Envoy proxy
	@cd envoy; docker run --name=envoy -p 8080:8080 --net=host ag_envoy

protoset:
	@echo Compiling protoset for grpcurl
	@cd ag; protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf \
	--proto_path=. --descriptor_set_out=ag.protoset --include_imports ag.proto

# change commands to 'go' when v.13 hits
test:
	@cd ./web; go1.13beta1 test
	@cd ./database; go1.13beta1 test
