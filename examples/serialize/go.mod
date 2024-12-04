module github.com/cvhariharan/plugin/example/serialize

go 1.23.3

replace github.com/cvhariharan/plugin => ../../

replace github.com/cvhariharan/plugin/example/serialize/plugin => ./plugin

require (
	github.com/cvhariharan/plugin v0.0.0-00010101000000-000000000000
	github.com/cvhariharan/plugin/example/serialize/plugin v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/lithammer/shortuuid v3.0.0+incompatible // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.68.0 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)
