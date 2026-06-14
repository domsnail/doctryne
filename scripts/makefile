PROTO_DIR=.
GO_OUT=../gen/go
TS_OUT=../gen/ts
GW_OPENAPI_OUT=../gen

all: protoc_gen_go protoc_gen_go_gw
dump_variables:
	echo "Using go mod path: ${GOPATH}" && \
	echo "Using protoc include path: ${PROTOC_INC}" && \
	echo "Using GRPC gateway include path: ${PROTO_GRPC_GW_INC}"

protoc_gen_go:
	cd ../api/proto && protoc -I . -I $(PROTOC_INC) -I $(PROTO_GRPC_GW_INC) \
		--go_out=$(GO_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT) \
		--go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
		inspection/v1/*.proto \
		utils/*.proto

# https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_openapi_output/

protoc_gen_go_gw:
	cd ../api/proto && protoc -I . -I $(PROTOC_INC) -I $(PROTO_GRPC_GW_INC) \
		--grpc-gateway_out=$(GO_OUT) \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt use_opaque_api=true \
		--openapiv2_out=$(GW_OPENAPI_OUT) \
		--openapiv2_opt=allow_merge=true \
		inspection/v1/*.proto \
		utils/v1/*.proto

protoc_gen_ts:
	cd ../api/proto && protoc -I . -I $(PROTO_DIR) -I $(PROTOC_INC) -I $(PROTO_GRPC_GW_INC) \
		--ts_proto_out=$(TS_OUT) \
		--ts_proto_opt=esModuleInterop=true,forceLong=string,outputServices=grpc-js \
		inspection/v1/*.proto \
		utils/v1/*.proto \
		scan/v1/*.proto

clean:
	del /S /Q $(GO_OUT)\* || echo ""
	del /S /Q $(TS_OUT)\* || echo ""

.PHONY: all go gateway ts clean
