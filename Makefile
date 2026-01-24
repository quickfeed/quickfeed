# This trick allow us to source the environment variables defined in .env file in the Makefile
# (see include directive in GNU make for more details).
# It ignores errors in case the .env file does not exists.
# It may be necessary to skip variables that uses special makefile characters, like $.
-include .env

OS			:= $(shell echo $(shell uname -s) | tr A-Z a-z)
github_user	:= $(shell gh auth status 2>/dev/null | awk '/Logged in to/ {print $$(NF-1)}')
dev_db		:= ./testdata/db/qf.db
protopatch	:= qf/types.proto kit/score/score.proto
proto_ts	:= $(protopatch:%.proto=public/proto/%_pb.ts)

# necessary when target is not tied to a specific file
.PHONY: download brew version-check dev-db install ui proto test qcm

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

dev-db:
	@if [ ! -f $(dev_db) ]; then \
		echo "Error: Database file not found at $(dev_db)"; \
		echo "Please download the database file from the QuickFeed organization."; \
		exit 1; \
	fi
	@echo "Updating development database with GitHub user: $(github_user) as QuickFeed admin"
	@python3 cmd/anonymize/main.py --database $(dev_db) --admin $(github_user)

install:
	@echo go install
	@go install
ifeq ($(OS),linux)
	@echo "Setting privileged ports capabilities for quickfeed"
	@sudo setcap 'cap_net_bind_service=+ep' `which quickfeed`
endif

ui: version-check
	@echo "Running npm ci and esbuild"
	@cd public; npm ci
	@go run cmd/esbuild/main.go

ui-update: version-check
	@echo "Running npm install and esbuild"
	@cd public; npm i
	@go run cmd/esbuild/main.go

overmind:
	@echo "Running Overmind Devtools"
	@cd public; npm run overmind

proto:
	buf dep update
	buf generate --template buf.gen.yaml

proto-swift:
	buf generate --template buf.gen.swift.yaml --exclude-path patch

test:
	@go clean -testcache
	@go test ./...
	@cd public && npm run test

qcm:
	@cd cmd/qcm; go install

clean-dev:
	# This is useful for debugging development environment and configuration issues.
	# WARNING: This will remove your .env file and configuration directory ~/.config/quickfeed
	@echo "Removing .env file and .env.bak file"
	@rm -f .env .env.bak
	@echo "Removing configuration directory ~/.config/quickfeed"
	@rm -rf ~/.config/quickfeed
	@echo "DOMAIN=localhost" > .env
	@echo ".env file created with DOMAIN=localhost"
