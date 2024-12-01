package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cvhariharan/plugin/catalog/protogen"
	"github.com/cvhariharan/plugin/store"
	"github.com/lithammer/shortuuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	PLUGIN_DISCOVERY_ADDRESS = "PLUGIN_DISCOVERY_ADDRESS"
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
	Address   string
	PluginMap map[string]Plugin
}

type PluginServeOptions struct {
	Name string
	Host string
}

type PluginResponse struct {
	SocketType string `json:"socket_type"`
	Address    string `json:"address"`
}

// Load loads a plugin either from a remote address or a local process.
// If the address is provided, it connects to the remote plugin using gRPC.
// If not, it starts the plugin in a subprocess and returns the client.
func Load(opt PluginLoadOptions, cs store.CatalogStore) (interface{}, error) {
	if opt.Address != "" {
		return loadRemote(opt)
	}

	return loadProcess(opt, cs)
}

// loadRemote connects to a remote plugin using gRPC and returns the client
func loadRemote(opt PluginLoadOptions) (interface{}, error) {
	conn, err := grpc.Dial(opt.Address, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("error connecting to remote plugin: %v", err)
	}

	p := opt.PluginMap[opt.Name]

	return p.Client(conn)
}

// loadProcess starts the plugin in a subprocess and returns the client
func loadProcess(opt PluginLoadOptions, cs store.CatalogStore) (interface{}, error) {
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

	var pluginResp PluginResponse
	scanner := bufio.NewScanner(stdout)
	if scanner.Scan() {
		resp := scanner.Text()
		if err := json.Unmarshal([]byte(resp), &pluginResp); err != nil {
			return nil, fmt.Errorf("error parsing plugin response: %v", err)
		}
	}

	if pluginResp.Address == "" {
		return nil, fmt.Errorf("plugin did not provide a valid address")
	}

	p := opt.PluginMap[opt.Name]

	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("unix", pluginResp.Address)
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithContextDialer(dialer),
	}

	conn, err := grpc.Dial("", opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create grpc client conn to %s: %v", pluginResp.Address, err)
	}

	if !cs.Add(opt.Name, store.ServiceInfo{
		Address: pluginResp.Address,
		Socket:  SOCKET_TYPE_UNIX,
	}) {
		return nil, fmt.Errorf("could not add service %s to catalog store", opt.Name)
	}

	return p.Client(conn)
}

// Serve starts a gRPC server for the plugin and listens on the provided address.
// If host is empty, it binds to all interfaces but returns the first non-loopback local IP address for the client and discovery server.
func Serve(p Plugin, opt PluginServeOptions) error {
	socketType := os.Getenv(PLUGIN_SOCKET_TYPE)

	var lis net.Listener
	var err error
	var reqSocket protogen.SocketType
	var resp PluginResponse
	resp.SocketType = socketType
	switch socketType {
	case SOCKET_TYPE_TCP:
		reqSocket = protogen.SocketType_TCP
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

		var localIP string
		if len(opt.Host) > 0 {
			localIP = opt.Host
		} else {
			localIP, err = getLocalIP()
			if err != nil {
				return err
			}
		}

		u, err := url.Parse(fmt.Sprintf("%s://%s:%d", reqSocket, localIP, lis.Addr().(*net.TCPAddr).Port))
		if err != nil {
			return fmt.Errorf("could not get plugin address: %v", err)
		}
		resp.Address = u.String()

	case SOCKET_TYPE_UNIX:
		reqSocket = protogen.SocketType_UNIX
		lis, err = getUnixSocket()
		if err != nil {
			return err
		}
		defer lis.Close()
		resp.Address = lis.Addr().String()
	}

	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		return fmt.Errorf("error encoding plugin response: %v", err)
	}

	// If PLUGIN_DISCOVERY_ADDRESS is set, register the plugin to the discovery server
	if len(os.Getenv(PLUGIN_DISCOVERY_ADDRESS)) != 0 {
		discoveryAddress := os.Getenv(PLUGIN_DISCOVERY_ADDRESS)
		listener, err := grpc.Dial(discoveryAddress, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("could not connect to discovery server: %v", err)
		}
		defer listener.Close()

		client := protogen.NewCatalogClient(listener)
		req := &protogen.Service{
			Name:       opt.Name,
			Address:    resp.Address,
			SocketType: reqSocket,
		}

		if _, err := client.Add(context.Background(), req); err != nil {
			return fmt.Errorf("could not register plugin to discovery server: %v", err)
		}
	}

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

// getLocalIP returns the local IP address of the machine.
// It iterates over all network interfaces and returns the first non-loopback IPv4 address it finds
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if !ip.IsLoopback() && ip.To4() != nil {
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("unable to find valid local IP address")
}
