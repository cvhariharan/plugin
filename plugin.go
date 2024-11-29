package plugin

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/lithammer/shortuuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	PLUGIN_DISCOVERY_ADDRESS = "PLUGIN_DISOVERY_ADDRESS"
	PLUGIN_SOCKET_TYPE       = "PLUGIN_SOCKET_TYPE"
	PLUGIN_MIN_PORT          = "PLUGIN_MIN_PORT"
	PLUGIN_MAX_PORT          = "PLUGIN_MAX_PORT"
	MIN_PORT                 = 10000
	MAX_PORT                 = 15000

	SOCKET_TYPE_TCP  = "tcp"
	SOCKET_TYPE_UNIX = "unix"
)

type Plugin interface {
	Client(*grpc.ClientConn) (interface{}, error)
	Server(*grpc.Server) error
}

type PluginLoadOptions struct {
	Name      string
	Path      string
	PluginMap map[string]Plugin
}

// Load starts the plugin in an exec.Command and returns the client
// to interact with the plugin. This client should implement the main
// business login interface and will internally route the RPC to the server
// using a gRPC client.
//
// This will create the client using a pluginMap to retrieve the plugin
// and running plugin.Client
func Load(opt PluginLoadOptions) (interface{}, error) {
	cmd := exec.Command(opt.Path)
	var env []string
	env = append(env, fmt.Sprintf("%s=unix", PLUGIN_SOCKET_TYPE))
	env = append(env, fmt.Sprintf("%s=%d", PLUGIN_MIN_PORT, MIN_PORT))
	env = append(env, fmt.Sprintf("%s=%d", PLUGIN_MAX_PORT, MAX_PORT))
	cmd.Env = append(cmd.Env, env...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not launch plugin %s: %v", opt.Path, err)
	}

	var socketPath string
	scanner := bufio.NewScanner(stdout)
	if scanner.Scan() {
		socketPath = scanner.Text()
	}

	p := opt.PluginMap[opt.Name]

	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(dialer),
	}

	conn, err := grpc.Dial("", opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create grpc client conn to %s: %v", socketPath, err)
	}

	return p.Client(conn)
}

// Serve starts serving a plugin over gRPC.
// It calls the Server method of the plugin to associate the gRPC server.
//
// Serve reads the config from env variables and automatically registers the plugin
// to the provided discovery server.
func Serve(p Plugin) error {
	socketType := os.Getenv(PLUGIN_SOCKET_TYPE)

	var lis net.Listener
	var err error
	switch socketType {
	case SOCKET_TYPE_TCP:
		min, err := strconv.Atoi(os.Getenv(PLUGIN_MIN_PORT))
		if err != nil {
			return fmt.Errorf("PLUGIN_MIN_PORT could not be parsed into int: %v", err)
		}

		max, err := strconv.Atoi(os.Getenv(PLUGIN_MAX_PORT))
		if err != nil {
			return fmt.Errorf("PLUGIN_MAX_PORT could not be parsed into int: %v", err)
		}

		lis, err = getTCPPort(min, max)
		if err != nil {
			return err
		}
		defer lis.Close()

	case SOCKET_TYPE_UNIX:
		lis, err = getUnixSocket()
		if err != nil {
			return err
		}
		defer lis.Close()
	}
	fmt.Println(lis.Addr().String())

	srv := getGRPCServer()
	p.Server(srv)
	return srv.Serve(lis)
}

// getTCPPort interates over the port range and finds an unused TCP port
func getTCPPort(min, max int) (net.Listener, error) {
	if min > max {
		return nil, fmt.Errorf("min port cannot be greater than max port")
	}

	for i := min; i < max; i++ {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", i))
		if err == nil {
			return lis, nil
		}
	}
	return nil, fmt.Errorf("no ports available")
}

func getUnixSocket() (net.Listener, error) {
	socketName := filepath.Join("/tmp", shortuuid.New())

	lis, err := net.Listen("unix", socketName)
	if err != nil {
		return nil, fmt.Errorf("could not listen on unix socket %s: %v", socketName, err)
	}

	return lis, err
}

// getGRPCServer returns a server with default values
func getGRPCServer() *grpc.Server {
	var opt []grpc.ServerOption
	grpcServer := grpc.NewServer(opt...)
	reflection.Register(grpcServer)

	return grpcServer
}
