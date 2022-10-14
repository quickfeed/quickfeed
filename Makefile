# This trick allow us to source the environment variables defined in .env file in the Makefile
# (see include directive in GNU make for more details).
# It ignores errors in case the .env file does not exists.
# It may be necessary to skip variables that uses special makefile caracters, like $.
-include .env

OS					:= $(shell echo $(shell uname -s) | tr A-Z a-z)
ARCH				:= $(shell uname -m)
proto-swift-path	:= ../quickfeed-swiftui/Quickfeed/Proto
grpcweb-latest		:= $(shell git ls-remote --tags https://github.com/grpc/grpc-web.git | tail -1 | awk -F"/" '{ print $$3 }')
grpcweb-ver			:= $(shell cd public; npm ls --package-lock-only grpc-web | awk -F@ '/grpc-web/ { print $$2 }')
protoc-grpcweb		:= protoc-gen-grpc-web
protoc-grpcweb-long	:= $(protoc-grpcweb)-$(grpcweb-ver)-$(OS)-$(ARCH)
sedi				:= $(shell sed --version >/dev/null 2>&1 && echo "sed -i --" || echo "sed -i ''")
toolsdir			:= bin
tool-pkgs			:= $(shell go list -f '{{join .Imports " "}}' tools.go)
tool-cmds			:= $(foreach tool,$(notdir ${tool-pkgs}),${toolsdir}/${tool}) $(foreach cmd,${tool-cmds},$(eval $(notdir ${cmd})Cmd := ${cmd}))

# necessary when target is not tied to a specific file
.PHONY: devtools download tools grpcweb install ui proto scm version-check

devtools: grpcweb tools

download:
	@echo "Download go.mod dependencies"
	@go mod download

go.mod: tools.go
	go mod tidy
	touch go.mod

${tool-cmds}: go.mod
	go build -o $@ $(filter %/$(@F),${tool-pkgs})

tools: ${tool-cmds}

version-check:
	@go run cmd/vercheck/main.go
ifneq ($(grpcweb-ver), $(grpcweb-latest))
	@echo WARNING: grpc-web version is not latest: $(grpcweb-ver) != $(grpcweb-latest)
endif

grpcweb:
	@echo "Fetch and install grpcweb protoc plugin"
	@mkdir -p $(toolsdir)
	@cd $(toolsdir); gh release download --repo grpc/grpc-web $(grpcweb-ver) --pattern \*$(OS)\*
	@cd $(toolsdir); shasum -c *.sha256 && rm *.sha256
	@cd $(toolsdir); mv $(protoc-grpcweb-long) $(protoc-grpcweb) && chmod +x $(protoc-grpcweb)

install:
	@echo go install
	@go install

define proto_target
proto_$(1):
	$$(info Compiling proto definitions for Go and TypeScript for $(1))
	@protoc --fatal_warnings -I . \
	-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
	-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
	--go-patch_out=plugin=go,paths=source_relative:. \
	--go-patch_out=plugin=go-grpc,paths=source_relative:. \
	--js_out=import_style=commonjs:$(1)/proto \
	--grpc-web_out=import_style=typescript,mode=grpcweb:$(1)/proto \
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
$(foreach dir,$(dirs),$(eval $(call proto_target,$(dir))))

ui: version-check
	@echo "Running npm ci and webpack"
	@cd public; npm ci; webpack

ui-update: version-check
	@echo "Running npm install and webpack"
	@cd public; npm i; webpack

proto:
	buf generate --template buf.gen.ui.yaml --exclude-path patch
	buf generate --template buf.gen.yaml

	@echo "Removing unused protopatch imports (see https://github.com/grpc/grpc-web/issues/529))"
	@$(sedi) '/patch_go_pb/d' \
	public/proto/kit/score/score_pb.js \
	public/proto/kit/score/score_pb.d.ts \
	public/proto/qf/types_pb.js \
	public/proto/qf/types_pb.d.ts

	@echo "Compiling proto files for frontend"
	@cd public && npm run tsc -- proto/qf/QuickfeedServiceClientPb.ts

# proto: proto_public

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
	@brew install gh go protobuf node docker clang-format golangci-lint bufbuild/buf/buf grpcurl

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

qcm:
	@cd cmd/qcm; go install

scm:
	@echo "Compiling the scm tool"
	@cd cmd/scm; go install
