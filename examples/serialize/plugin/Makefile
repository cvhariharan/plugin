.PHONY: protoc

protoc:
	protoc --go_out=. --go_opt=module=plugin/example/serialize \
    --go-grpc_out=. --go-grpc_opt=module=plugin/example/serialize \
    protos/*.proto