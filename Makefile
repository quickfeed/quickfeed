# This trick allow us to source the environment variables defined in .env file in the Makefile 
# (see include directive in GNU make for more details).
# It ignores errors in case the .env file does not exists.
# It may be necessary to skip variables that uses special makefile caracters, like $.
-include .env

OS					:= $(shell echo $(shell uname -s) | tr A-Z a-z)
ARCH				:= $(shell uname -m)
tmpdir				:= tmp
ui					:= dev
proto-path			:= $(ui)/proto
proto-swift-path	:= ../quickfeed-swiftui/Quickfeed/Proto
grpcweb-ver			:= 1.3.1
protoc-grpcweb		:= protoc-gen-grpc-web
protoc-grpcweb-long	:= $(protoc-grpcweb)-$(grpcweb-ver)-$(OS)-$(ARCH)
grpcweb-url			:= https://github.com/grpc/grpc-web/releases/download/$(grpcweb-ver)/$(protoc-grpcweb-long)
grpcweb-path		:= /usr/local/bin/$(protoc-grpcweb)
sedi				:= $(shell sed --version >/dev/null 2>&1 && echo "sed -i --" || echo "sed -i ''")
testorg				:= ag-test-course
envoy-config-gen	:= ./cmd/envoy/envoy_config_gen.go

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
	@echo "Fetch and install grpcweb protoc plugin (may require sudo access on some systems)"
	@mkdir -p $(tmpdir)
	@cd $(tmpdir); curl -LOs $(grpcweb-url)
	@sudo mv $(tmpdir)/$(protoc-grpcweb-long) $(grpcweb-path)
	@chmod +x $(grpcweb-path)
	@rm -rf $(tmpdir)

install:
	@echo go install
	@go install

ui:
	@echo Running npm ci and webpack for $(ui)
	@cd $(ui); npm ci; webpack

proto:
	@echo "Compiling QuickFeed's ag and kit/score proto definitions for Go and TypeScript"
	@protoc \
	-I . \
	-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
	-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
	--go-patch_out=plugin=go,paths=source_relative:. \
	--go-patch_out=plugin=go-grpc,paths=source_relative:. \
	--js_out=import_style=commonjs:$(proto-path) \
	--grpc-web_out=import_style=typescript,mode=grpcwebtext:$(proto-path) \
	ag/ag.proto kit/score/score.proto

	@echo "Removing unused protopatch imports (see https://github.com/grpc/grpc-web/issues/529)"
	@$(sedi) '/patch_go_pb/d' \
	$(proto-path)/kit/score/score_pb.js \
	$(proto-path)/kit/score/score_pb.d.ts \
	$(proto-path)/ag/ag_pb.js \
	$(proto-path)/ag/ag_pb.d.ts \
	$(proto-path)/ag/AgServiceClientPb.ts
	@echo "Compiling proto for $(ui)"
	@cd $(ui) && npm run tsc -- proto/ag/AgServiceClientPb.ts

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
	$(error "No brew command in $(PATH)")
endif
	@echo "Installing homebrew packages needed for development and deployment"
	@brew install go protobuf webpack npm node docker certbot envoy

envoy-config:
ifeq ($(DOMAIN),)
	@echo "You must set required environment variables before configuring Envoy (see .env-template)." && false
else
	@echo "Generating Envoy configuration for $(DOMAIN)."
	@go run $(envoy-config-gen) --tls
endif

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

run:
	@quickfeed -service.url $(DOMAIN) -database.file ./tmp.db

runlocal:
	@quickfeed -service.url 127.0.0.1

prometheus:
	sudo prometheus --web.listen-address="localhost:9095" --config.file=metrics/prometheus.yml --storage.tsdb.path=/var/lib/prometheus/data --storage.tsdb.retention.size=1024MB --web.external-url=http://localhost:9095/stats --web.route-prefix="/" &
