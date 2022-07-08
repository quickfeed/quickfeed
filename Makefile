# This trick allow us to source the environment variables defined in .env file in the Makefile
# (see include directive in GNU make for more details).
# It ignores errors in case the .env file does not exists.
# It may be necessary to skip variables that uses special makefile caracters, like $.
-include .env

OS					:= $(shell echo $(shell uname -s) | tr A-Z a-z)
ARCH				:= $(shell uname -m)
tmpdir				:= tmp
proto-swift-path	:= ../quickfeed-swiftui/Quickfeed/Proto
grpcweb-latest		:= $(shell git ls-remote --tags https://github.com/grpc/grpc-web.git | tail -1 | awk -F"/" '{ print $$3 }')
grpcweb-ver			:= $(shell cd public; npm ls --package-lock-only grpc-web | awk -F@ '/grpc-web/ { print $$2 }')
protoc-grpcweb		:= protoc-gen-grpc-web
protoc-grpcweb-long	:= $(protoc-grpcweb)-$(grpcweb-ver)-$(OS)-$(ARCH)
grpcweb-url			:= https://github.com/grpc/grpc-web/releases/download/$(grpcweb-ver)/$(protoc-grpcweb-long)
grpcweb-path		:= /usr/local/bin/$(protoc-grpcweb)
sedi				:= $(shell sed --version >/dev/null 2>&1 && echo "sed -i --" || echo "sed -i ''")
testorg				:= ag-test-course
envoy-config-gen	:= ./cmd/envoy/envoy_config_gen.go

# necessary when target is not tied to a file
.PHONY: devtools download go-tools grpcweb install ui proto envoy-build envoy-run scm version-check

devtools: grpcweb go-tools

download:
	@echo "Download go.mod dependencies"
	@go mod download

go-tools:
	@echo "Installing tools from tools.go"
	@go install `go list -f "{{range .Imports}}{{.}} {{end}}" tools.go`

version-check:
	@go run cmd/vercheck/main.go
ifneq ($(grpcweb-ver), $(grpcweb-latest))
	@echo WARNING: grpc-web version is not latest: $(grpcweb-ver) != $(grpcweb-latest)
endif

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

define ui_target
ui_$(1):
	$$(info Running npm ci and webpack for $(1))
	@cd $(1); npm ci; webpack
endef

define proto_target
proto_$(1):
	$$(info Compiling proto definitions for Go and TypeScript for $(1))
	@protoc --fatal_warnings -I . \
	-I ./proto-include \
	-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
	-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
	--go-patch_out=plugin=go,paths=source_relative:. \
	--go-patch_out=plugin=go-grpc,paths=source_relative:. \
	--js_out=import_style=commonjs:$(1)/proto \
	--grpc-web_out=import_style=typescript,mode=grpcwebtext:$(1)/proto \
	qf/quickfeed.proto qf/types.proto qf/requests.proto kit/score/score.proto

	$$(info Removing unused protopatch imports (see https://github.com/grpc/grpc-web/issues/529))
	@$(sedi) '/patch_go_pb/d' \
	$(1)/proto/kit/score/score_pb.js \
	$(1)/proto/kit/score/score_pb.d.ts \
	$(1)/proto/qf/quickfeed_pb.js \
	$(1)/proto/qf/quickfeed_pb.d.ts \
	$(1)/proto/qf/QuickfeedServiceClientPb.ts \
	$(1)/proto/qf/types_pb.js \
	$(1)/proto/qf/types_pb.d.ts \
	$(1)/proto/qf/requests_pb.js \
	$(1)/proto/qf/requests_pb.d.ts

	$$(info Compiling proto for $(1))
	@cd $(1) && npm run tsc -- proto/qf/QuickfeedServiceClientPb.ts
endef

dirs := public
$(foreach dir,$(dirs),$(eval $(call ui_target,$(dir))))
$(foreach dir,$(dirs),$(eval $(call proto_target,$(dir))))

ui: version-check ui_public

proto: proto_public

proto-swift:
	@echo "Compiling QuickFeed's proto definitions for Swift"
	@protoc \
	-I . \
	-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
	-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
	--swift_out=:$(proto-swift-path) \
	--grpc-swift_out=$(proto-swift-path) \
	qf/quickfeed.proto

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
	--proto_path=qf \
	--descriptor_set_out=qf/qf.protoset \
	--include_imports \
	qf/quickfeed.proto

test:
	@go clean -testcache ./...
	@go test ./...

webpack-dev-server:
	@cd public && npx webpack-dev-server --config webpack.config.js --port 8082 --progress --mode development

# TODO Should check that webpack-dev-server is running.
selenium:
	@cd public && npm run test:selenium

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
