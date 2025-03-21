# This trick allow us to source the environment variables defined in .env file in the Makefile
# (see include directive in GNU make for more details).
# It ignores errors in case the .env file does not exists.
# It may be necessary to skip variables that uses special makefile characters, like $.
-include .env

OS					:= $(shell echo $(shell uname -s) | tr A-Z a-z)
ARCH				:= $(shell uname -m)

# necessary when target is not tied to a specific file
.PHONY: download brew version-check install ui proto test qcm cm

download:
	@echo "Download go.mod dependencies"
	@go mod download

brew:
ifeq (, $(shell which brew))
	$(error "No brew command in $(PATH)")
endif
	@echo "Installing homebrew packages needed for development and deployment"
	@brew install gh go protobuf node docker clang-format golangci-lint bufbuild/buf/buf grpcurl goreleaser

version-check:
	@go run cmd/vercheck/main.go

install:
	@echo go install
	@go install
ifeq ($(OS),linux)
	@echo "Setting privileged ports capabilities for quickfeed"
	@sudo setcap 'cap_net_bind_service=+ep' `which quickfeed`
endif

ui: version-check
	@echo "Running npm ci and webpack"
	@cd public; npm ci; webpack

ui-update: version-check
	@echo "Running npm install and webpack"
	@cd public; npm i; webpack

proto:
	buf dep update
	buf generate --template buf.gen.yaml

# TODO(meling): Split the proto target to avoid generating too new typescript... Need to fix #1147 first; after which we should merge this target with the proto target.
proto-ui: $(protopatch)
	buf generate --template buf.gen.ui.yaml --exclude-path patch

proto-swift:
	buf generate --template buf.gen.swift.yaml --exclude-path patch

test:
	@go clean -testcache
	@go test ./...

webpack-dev-server:
	@cd public && npx webpack-dev-server --config webpack.config.js --port 8082 --progress --mode development

# TODO Should check that webpack-dev-server is running.
selenium:
	@cd public && npm run test:selenium

qcm:
	@cd cmd/qcm; go install

cm:
	@go install github.com/quickfeed/quickfeed/cmd/cm
