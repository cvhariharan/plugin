package shared

import (
	"context"
	"plugin"

	"github.com/cvhariharan/plugin/example/hello/plugin/protos"
	"google.golang.org/grpc"
)

// This is the main business logic interface
type Hello interface {
	Greet() string
}

type HelloPlugin struct {
	plugin.Plugin

	// This is the implementation that will be called by the server
	Impl Hello
}

func (h *HelloPlugin) Client(conn *grpc.ClientConn) (interface{}, error) {
	c := protos.NewHelloClient(conn)
	return &HelloClient{client: c}, nil
}

func (h *HelloPlugin) Server(srv *grpc.Server) error {
	hs := &HelloServer{Impl: h.Impl}
	protos.RegisterHelloServer(srv, hs)
	return nil
}

// This client will implement the business logic interface
// and internally use a gRPC client to talk to the server.
type HelloClient struct {
	client protos.HelloClient
}

func (hc *HelloClient) Greet() string {
	resp, err := hc.client.Greet(context.Background(), &protos.Empty{})
	if err != nil {
		return ""
	}
	return resp.GetHello()
}

// This is the gRPC server that will internally call the actual implementation
type HelloServer struct {
	protos.UnimplementedHelloServer
	Impl Hello
}

func (hs *HelloServer) Greet(ctx context.Context, e *protos.Empty) (*protos.Resp, error) {
	r := hs.Impl.Greet()
	return &protos.Resp{Hello: r}, nil
}
