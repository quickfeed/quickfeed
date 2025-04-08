module github.com/quickfeed/quickfeed

go 1.24.2

require (
	connectrpc.com/connect v1.18.1
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/alecthomas/kong v1.9.0
	github.com/alta/protopatch v0.5.3
	github.com/beatlabs/github-auth v0.0.0-20240915075323-ab38acb2d2a1
        github.com/docker/docker v28.0.4+incompatible
        github.com/evanw/esbuild v0.25.2
	github.com/fsnotify/fsnotify v1.9.0
        github.com/go-git/go-git/v5 v5.14.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/go-cmp v0.7.0
	github.com/google/go-github/v62 v62.0.0
	github.com/prometheus/client_golang v1.21.1
	github.com/quickfeed/quickfeed/kit v0.11.0
	github.com/shurcooL/githubv4 v0.0.0-20240727222349-48295856cce7
	github.com/steinfletcher/apitest v1.6.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.37.0
	golang.org/x/net v0.38.0
	golang.org/x/oauth2 v0.29.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudflare/circl v1.6.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.27 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.63.0 // indirect
	github.com/prometheus/procfs v0.16.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shurcooL/graphql v0.0.0-20230722043721-ed46e5a46466 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)

replace github.com/quickfeed/quickfeed/kit => ./kit

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	github.com/alta/protopatch/cmd/protoc-gen-go-patch
	google.golang.org/protobuf/cmd/protoc-gen-go
)
