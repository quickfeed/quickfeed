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
testorg				:= ag-test-course
endpoint 			:= test.itest.run
agport				:= 8081
pbpath				:= $(shell go list -f '{{ .Dir }}' -m github.com/gogo/protobuf)

# necessary when target is not tied to a file
.PHONY: download install-tools install ui proto devtools grpcweb envoy-build envoy-run scm

devtools: install-tools grpcweb

download:
	@echo Download go.mod dependencies
	@go mod download

install-tools: download
	@echo Installing tools from tools.go
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

install:
	@echo go install
	@go install

ui:
	@echo Running webpack
	@cd public; npm install; webpack

proto:
	@echo Compiling Autograders proto definitions
	@cd ag; protoc -I=. -I=$(pbpath) --gogofast_out=plugins=grpc,\
	Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
	Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:. \
	--js_out=import_style=commonjs:../$(proto-path)/ \
	--grpc-web_out=import_style=typescript,mode=grpcweb:../$(proto-path)/ ag.proto
	$(sedi) '/gogo/d' $(proto-path)/ag_pb.js $(proto-path)/AgServiceClientPb.ts $(proto-path)/ag_pb.d.ts
	@tsc $(proto-path)/AgServiceClientPb.ts

grpcweb:
	@echo "Fetch and install grpcweb protoc plugin (requires sudo access)"
	@mkdir -p $(tmpdir)
	@cd $(tmpdir); curl -LOs $(grpcweb-url)
	@sudo mv $(tmpdir)/$(protoc-grpcweb-long) $(grpcweb-path)
	@chmod +x $(grpcweb-path)
	@rm -rf $(tmpdir)

# TODO(meling) this is just for macOS; we should guard against non-macOS.
brew:
	@echo "Install homebrew packages needed for development"
	@brew update
	@brew cleanup
	@brew install go protobuf npm docker

envoy-build:
	@echo "Building Autograder Envoy proxy"
	@cd envoy; docker build -t ag_envoy -f envoy.Dockerfile .

envoy-run:
	@echo "Starting Autograder Envoy proxy"
	@cd envoy; docker run --name=envoy -p 8080:8080 --net=host ag_envoy

# will stop envoy container, prune docker containers and remove envoy images
# use before rebuilding envoy with changed configuration in envoy.yaml
envoy-purge:
	@docker kill envoy
	@docker container prune
	@docker image rm envoyproxy/envoy ag_envoy

# protoset is a file used as a server reflection to mock-testing of grpc methods via command line
protoset:
	@echo "Compiling protoset for grpcurl"
	@cd ag; protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf \
	--proto_path=. --descriptor_set_out=ag.protoset --include_imports ag.proto

test:
	@go clean -testcache ./...
	@go test ./...

scm:
	@echo "Compiling the scm tool"
	@cd cmd/scm; go install

# will remove all repositories and teams from provided organization 'testorg'
purge: scm
	@scm delete repo -all -namespace=$(testorg)
	@scm delete team -all -namespace=$(testorg)

# will start ag client and server, serve static files at 'endpoint' and webserver at 'agport'
# change agport variable to the number of bound local port when using tunnel script
run:
	@quickfeed -service.url $(endpoint) -http.addr :$(agport) -http.public ./public -database.file ./tmp.db

runlocal:
	@quickfeed -service.url 127.0.0.1 -http.addr :9091 -http.public ./public

# test nginx configuration syntax
nginx-test:
	@sudo nginx -t

# restart nginx with updated configuration
nginx: nginx-test
	@sudo nginx -s reload

# changes where the grpc-client is being run, use "remote" target when starting from ag2
local:
	@echo "Changing grpc client location to localhost"
	@cd ./public/src/managers/; sed -i 's/"https:\/\/" + window.location.hostname/"http:\/\/localhost:8080"/g' GRPCManager.ts
	@cd ./public; webpack

remote:
	@echo "Changing grpc client location to remote domain"
	@cd ./public/src/managers/; sed -i 's/"http:\/\/localhost:8080"/"https:\/\/" + window.location.hostname/g' GRPCManager.ts

prometheus:
	sudo prometheus --web.listen-address="localhost:9095" --config.file=metrics/prometheus.yml --storage.tsdb.path=/var/lib/prometheus/data --storage.tsdb.retention.size=1024MB --web.external-url=http://localhost:9095/stats --web.route-prefix="/" &

quickfeed-go:
	docker build -f ci/scripts/go/Dockerfile -t quickfeed:go .
