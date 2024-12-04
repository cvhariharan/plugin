package shared

import (
	"context"
	"fmt"
	"log"

	pb "github.com/cvhariharan/plugin/example/serialize/plugin/protos"

	"github.com/cvhariharan/plugin"
	"google.golang.org/grpc"
)

type Test interface {
	TestCall(*TestObj) string
}

type TestPlugin struct {
	plugin.Plugin
	Impl Test
}

func (tp *TestPlugin) Client(conn *grpc.ClientConn) (interface{}, error) {
	c := pb.NewTestClient(conn)
	return &TestClient{client: c}, nil
}

func (tp *TestPlugin) Server(srv *grpc.Server) error {
	ts := &TestServer{Impl: tp.Impl}
	pb.RegisterTestServer(srv, ts)
	return nil
}

type TestClient struct {
	client pb.TestClient
}

func (tc *TestClient) TestCall(o *TestObj) string {
	b, t, err := plugin.SerializeObject(o)
	if err != nil {
		log.Println(err)
		return ""
	}

	r, err := tc.client.TestCall(context.Background(), &pb.Obj{SerializedObjects: b, TypeName: t})
	if err != nil {
		log.Println(err)
		return ""
	}

	return r.Response
}

type TestServer struct {
	pb.UnimplementedTestServer
	Impl Test
}

func (ts *TestServer) TestCall(ctx context.Context, obj *pb.Obj) (*pb.Resp, error) {
	rObj, err := plugin.DeserializeObject(obj.SerializedObjects, obj.TypeName)
	if err != nil {
		return nil, fmt.Errorf("could not deserialize object: %v", err)
	}

	if t, ok := rObj.(*TestObj); ok {
		return &pb.Resp{Response: t.TestCall()}, nil
	}

	return nil, nil
}
