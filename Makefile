OS					:= $(shell echo $(shell uname -s) | tr A-Z a-z)
ARCH				:= $(shell uname -m)
tmpdir				:= tmp
proto-path			:= public/proto
proto-swift-path	:= ../quickfeed-swiftui/Quickfeed/Proto
grpcweb-ver			:= 1.2.0
protoc-grpcweb		:= protoc-gen-grpc-web
protoc-grpcweb-long	:= $(protoc-grpcweb)-$(grpcweb-ver)-$(OS)-$(ARCH)
grpcweb-url			:= https://github.com/grpc/grpc-web/releases/download/$(grpcweb-ver)/$(protoc-grpcweb-long)
grpcweb-path		:= /usr/local/bin/$(protoc-grpcweb)
sedi				:= $(shell sed --version >/dev/null 2>&1 && echo "sed -i --" || echo "sed -i ''")
testorg				:= ag-test-course
endpoint 			:= test.itest.run
agport				:= 8081

# necessary when target is not tied to a file
.PHONY: devtools download go-tools grpcweb install ui proto envoy-build envoy-run scm

devtools: grpcweb go-tools

download:
	@echo "Download go.mod dependencies"
	@go mod download

go-tools:
	@echo "Installing tools from tools.go"
	@go install `go list -f "{{range .Imports}}{{.}} {{end}}" tools.go`

grpcweb:
	@echo "Fetch and install grpcweb protoc plugin"
	@mkdir -p $(tmpdir)
	@cd $(tmpdir); curl -LOs $(grpcweb-url)
	@mv $(tmpdir)/$(protoc-grpcweb-long) $(grpcweb-path)
	@chmod +x $(grpcweb-path)
	@rm -rf $(tmpdir)

install:
	@echo go install
	@go install

ui:
	@echo Running webpack
	@cd public; npm install; npm run webpack

proto:
	@echo "Compiling QuickFeed's proto definitions for Go and TypeScript"
	@protoc \
	-I . \
	-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
	-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
	--go-patch_out=plugin=go,paths=source_relative:. \
	--go-patch_out=plugin=go-grpc,paths=source_relative:. \
	--js_out=import_style=commonjs:$(proto-path) \
	--grpc-web_out=import_style=typescript,mode=grpcwebtext:$(proto-path) \
	ag/ag.proto
	@echo "Removing unused protopatch imports (see https://github.com/grpc/grpc-web/issues/529)"
	@$(sedi) '/patch_go_pb/d' \
	$(proto-path)/ag/ag_pb.js \
	$(proto-path)/ag/ag_pb.d.ts \
	$(proto-path)/ag/AgServiceClientPb.ts
	@cd public && npm run tsc -- proto/ag/AgServiceClientPb.ts

proto-swift:
	@echo "Compiling QuickFeed's proto definitions for Swift"
	@protoc \
	-I . \
	-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
	-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
	--swift_out=:$(proto-swift-path) \
	--grpc-swift_out=$(proto-swift-path) \
	ag/ag.proto

brew:
    ifeq (, $(shell which brew))
		$(error "No brew command in $$PATH")
    endif
	@echo "Installing homebrew packages needed for development and deployment"
	@brew install go protobuf webpack npm node docker certbot envoy

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
	@protoc \
	-I . \
	-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
	-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
	--proto_path=ag \
	--descriptor_set_out=ag/ag.protoset \
	--include_imports \
	ag/ag.proto

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
	@cd ./public/src/managers/; $(sedi) 's/"https:\/\/" + window.location.hostname/"http:\/\/localhost:8080"/g' GRPCManager.ts
	@cd ./public; webpack

remote:
	@echo "Changing grpc client location to remote domain"
	@cd ./public/src/managers/; $(sedi) 's/"http:\/\/localhost:8080"/"https:\/\/" + window.location.hostname/g' GRPCManager.ts

envoy-config:
ifeq ($(DOMAIN),)
	@echo "You must set required environment variables before configuring Envoy (see doc/scripts/envs.sh)."
else
	@echo "Generating Envoy configuration for '$$DOMAIN'."
	@$(shell CONFIG='$$DOMAIN:$$GRPC_PORT:$$HTTP_PORT'; envsubst "$$CONFIG" < envoy/envoy.tmpl > $$ENVOY_CONFIG)
endif

prometheus:
	sudo prometheus --web.listen-address="localhost:9095" --config.file=metrics/prometheus.yml --storage.tsdb.path=/var/lib/prometheus/data --storage.tsdb.retention.size=1024MB --web.external-url=http://localhost:9095/stats --web.route-prefix="/" &

quickfeed-go:
	docker build -f ci/scripts/go/Dockerfile -t quickfeed:go .
