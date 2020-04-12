module github.com/autograde/aguis

go 1.14

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/autograde/aguis/kit v0.0.0-00010101000000-000000000000
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.5
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v30 v30.0.0
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gorilla/sessions v1.2.0
	github.com/gosimple/slug v1.9.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/labstack/echo-contrib v0.9.0
	github.com/labstack/echo/v4 v4.1.15
	github.com/markbates/goth v1.62.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/tebeka/selenium v0.9.9
	github.com/urfave/cli v1.22.2
	github.com/xanzy/go-gitlab v0.28.0
	go.uber.org/zap v1.14.1
	golang.org/x/crypto v0.0.0-20200320181102-891825fb96df // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200321134203-328b4cd54aae // indirect
	golang.org/x/tools v0.0.0-20200321224714-0d839f3cf2ed // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20200323114720-3f67cca34472 // indirect
	google.golang.org/grpc v1.28.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/webhooks.v3 v3.13.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v0.17.1
	k8s.io/metrics v0.17.1
	k8s.io/utils v0.0.0-20200327001022-6496210b90e8 // indirect
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
)

replace github.com/autograde/aguis/kit => ./kit
