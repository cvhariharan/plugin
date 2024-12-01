.PHONY: protoc

protoc:
	protoc --go_out=. --go_opt=module=github.com/cvhariharan/plugin \
    --go-grpc_out=. --go-grpc_opt=module=github.com/cvhariharan/plugin \
    catalog/protos/*.proto