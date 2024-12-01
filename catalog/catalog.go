package catalog

import (
	"context"
	"fmt"
	"net"

	"github.com/cvhariharan/plugin/catalog/protogen"
	"github.com/cvhariharan/plugin/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type CatalogServer struct {
	protogen.UnimplementedCatalogServer
	Impl store.CatalogStore
}

func Serve(cs store.CatalogStore, address string) error {
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	var opt []grpc.ServerOption
	srv := grpc.NewServer(opt...)

	c := &CatalogServer{
		Impl: cs,
	}
	protogen.RegisterCatalogServer(srv, c)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("could not start catalog server, could not listen on address: %w", err)
	}
	return srv.Serve(lis)
}

func (c *CatalogServer) Add(ctx context.Context, req *protogen.Service) (*protogen.Empty, error) {
	var socketType store.SocketType
	switch req.SocketType {
	case protogen.SocketType_TCP:
		socketType = store.TCP
	case protogen.SocketType_UNIX:
		socketType = store.UNIX
	default:
		return nil, fmt.Errorf("invalid socket type")
	}

	if ok := c.Impl.Add(req.Name, store.ServiceInfo{Address: req.Address, Socket: socketType}); !ok {
		return nil, fmt.Errorf("failed to add service %s", req.Name)
	}

	return &protogen.Empty{}, nil
}

func (c *CatalogServer) Get(ctx context.Context, req *protogen.GetReq) (*protogen.Service, error) {
	svcInfo, ok := c.Impl.Get(req.Name)
	if !ok {
		return nil, fmt.Errorf("service %s not found", req.Name)
	}

	var socketType protogen.SocketType
	switch svcInfo.Socket {
	case store.TCP:
		socketType = protogen.SocketType_TCP
	case store.UNIX:
		socketType = protogen.SocketType_UNIX
	default:
		return nil, fmt.Errorf("invalid socket type")
	}

	return &protogen.Service{
		Name:       req.Name,
		Address:    svcInfo.Address,
		SocketType: socketType,
	}, nil
}
