
.PHONY: install
install:
	@echo go install
	@go install

.PHONY: ui
ui:
	@echo running webpack
	@cd public; npm install; webpack

.PHONY: proto
proto:
	@echo compiling Autograders proto definitions
	@cd ag; protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf --gogofast_out=plugins=grpc,\
		Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
		Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
		Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
		Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
		Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:.\
		--js_out=import_style=commonjs:../public/proto/	\
		--grpc-web_out=import_style=typescript,mode=grpcweb:../public/proto/ ag.proto
	# @cd ag; sed -i '/gogo/d' ../public/proto/ag_pb.js ../public/proto/AgServiceClientPb.ts ../public/proto/ag_pb.d.ts
	# tsc public/proto/AgServiceClientPb.ts
