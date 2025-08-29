# QuickFeed Development Instructions

QuickFeed is a Go/TypeScript web application for automated feedback on programming assignments. It features a Go backend with gRPC/Connect services and a React/TypeScript frontend with Overmind state management.

**Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Working Effectively

### Dependencies and Environment Setup
- **Go**: Requires Go 1.24.1+ (compatible with 1.24.6)
- **Node.js**: Requires Node.js 20+ with npm
- **Protocol Buffers**: Requires buf, protoc, and Go protobuf tools
- **Docker**: Required for running tests with DOCKER_TESTS=1

Install Go tools (add `/home/runner/go/bin` to PATH):
```bash
export PATH=$PATH:/home/runner/go/bin
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
```

Install system dependencies (Ubuntu/Debian):
```bash
sudo apt-get update && sudo apt-get install -y protobuf-compiler
```

### Bootstrap and Build Process
**NEVER CANCEL builds or long-running commands. Set timeouts to 90+ minutes.**

1. **Download Go dependencies** - takes ~20 seconds:
```bash
make download
```

2. **Install frontend dependencies** - takes ~22 seconds:
```bash
cd public && npm ci
```

3. **Build Go backend** - takes ~52 seconds. NEVER CANCEL. Set timeout to 90+ minutes:
```bash
make install
```

4. **Build frontend UI** - takes ~4.5 seconds. NEVER CANCEL. Set timeout to 90+ minutes:
```bash
make ui
```

**Protocol buffer generation (optional)** - requires network access to buf.build:
```bash
make proto
```
Note: If `make proto` fails due to network issues (buf.build unavailable), the repository contains pre-generated protobuf files and builds will still work.

### Testing
**Run complete test suite** - takes ~1 min 33 seconds. NEVER CANCEL. Set timeout to 180+ minutes:
```bash
make test
```
This runs both Go tests and Jest frontend tests. Go tests are comprehensive; frontend tests may have 1-2 flaky date-related failures that are acceptable.

**Run only Go tests** - takes ~13 seconds:
```bash
go test ./...
```

**Run only frontend tests** - takes ~73 seconds:
```bash
cd public && npm run test
```

## Running QuickFeed

### Development Server Setup
Create `.env` file from template and configure for localhost:
```bash
cp .env-template .env
```

Generate required certificates and keys:
```bash
# Generate self-signed SSL certificates for HTTPS
mkdir -p internal/config/certs
openssl req -x509 -newkey rsa:2048 -keyout internal/config/certs/privkey.pem -out internal/config/certs/fullchain.pem -days 365 -nodes -subj "/CN=127.0.0.1"

# Generate GitHub app private key (for development)
mkdir -p internal/config/github
openssl genrsa -out internal/config/github/quickfeed.pem 2048
openssl rsa -in internal/config/github/quickfeed.pem -out internal/config/github/quickfeed.pem -traditional
```

Generate authentication secret:
```bash
quickfeed -secret
```

Update `.env` file for development:
```bash
# Set these values in .env
DOMAIN="127.0.0.1"
PORT="8080"
QUICKFEED_APP_ID="dev-app-id"
QUICKFEED_CLIENT_ID="dev-client-id"
QUICKFEED_CLIENT_SECRET="dev-client-secret"
# Comment out QUICKFEED_WHITELIST for localhost
```

### Start Development Server
```bash
export PATH=$PATH:/home/runner/go/bin
PORT=8080 quickfeed -dev
```
Server starts on https://127.0.0.1:8080 with self-signed certificates.

## Validation and Testing

### Manual Validation Requirements
**ALWAYS perform these validation steps after making changes:**

1. **Build validation**:
```bash
make download && make install && make ui
```

2. **Test validation**:
```bash
make test  # Must pass Go tests, 80+ frontend tests should pass
```

3. **Server functionality validation**:
```bash
PORT=8080 quickfeed -dev &
curl -k -s -o /dev/null -w "%{http_code}" https://127.0.0.1:8080/  # Should return 200
curl -k -s https://127.0.0.1:8080/ | grep "QuickFeed"  # Should show HTML with "QuickFeed"
```

4. **Frontend build artifacts validation**:
Check that `public/dist/` contains generated JavaScript and CSS files after `make ui`.

### Code Quality
**Always run before committing**:
```bash
cd public && npm run lint  # Frontend linting
# Go linting is handled by CI
```

## Important File Locations

### Core Application Structure
- `main.go` - Main server entry point
- `qf/` - Protocol buffer definitions and generated Go code
- `public/` - Frontend TypeScript/React application
- `public/src/` - Frontend source code
- `public/dist/` - Generated frontend build artifacts
- `web/` - Go HTTP handlers and web services
- `database/` - Database layer and models
- `internal/` - Internal Go packages

### Build and Configuration
- `Makefile` - Primary build orchestration
- `go.mod` - Go dependencies and module definition  
- `public/package.json` - Frontend dependencies and scripts
- `buf.gen.yaml` - Protocol buffer code generation config
- `.env` - Environment configuration (copy from .env-template)

### Development and Testing
- `doc/dev.md` - Developer guide with detailed setup instructions
- `doc/deploy.md` - Deployment guide with environment setup
- `testdata/` - Test data and fixtures
- `ci/` - Continuous integration and testing utilities

## Common Development Patterns

### After Editing Protocol Buffers (qf/*.proto)
```bash
make proto  # Regenerate protobuf code
make install && make ui  # Rebuild both backend and frontend
```

### After Frontend Changes (public/src/*)
```bash
make ui  # Rebuild frontend only
```

### After Go Backend Changes
```bash
make install  # Rebuild backend only
```

### Environment Troubleshooting
- If `buf` commands fail, ensure `/home/runner/go/bin` is in PATH
- If SSL errors occur, verify certificates in `internal/config/certs/`
- If GitHub key errors occur, verify RSA key in `internal/config/github/quickfeed.pem`
- If port binding fails, use non-privileged ports (8080+) or `sudo setcap 'cap_net_bind_service=+ep' $(which quickfeed)`

## Time Expectations
**NEVER CANCEL these operations. Always use generous timeouts:**

- Go dependency download: ~20 seconds
- Frontend dependency installation: ~22 seconds  
- Go backend build: ~52 seconds (timeout: 90+ minutes)
- Frontend build: ~4.5 seconds (timeout: 90+ minutes)
- Complete test suite: ~93 seconds (timeout: 180+ minutes)
- Protocol buffer generation: ~1-30 seconds (may fail with network issues)

These instructions ensure consistent, reliable development workflow for QuickFeed.