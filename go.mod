module github.com/autograde/quickfeed

go 1.14

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/alta/protopatch v0.3.4
	github.com/autograde/quickfeed/kit v0.0.0-20201031112539-d315b568016d
	github.com/containerd/containerd v1.5.2 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20200428022330-06a60b6afbbc // indirect
	github.com/docker/docker v20.10.7+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github/v35 v35.3.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/sessions v1.2.1
	github.com/gosimple/slug v1.9.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-hclog v0.12.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/now v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/labstack/echo-contrib v0.9.0
	github.com/labstack/echo/v4 v4.1.17
	github.com/lib/pq v1.5.1 // indirect
	github.com/markbates/goth v1.66.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.15.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/urfave/cli v1.22.4
	github.com/xanzy/go-gitlab v0.39.0
	go.uber.org/atomic v1.8.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.17.0
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/autograde/quickfeed/kit => ./kit
