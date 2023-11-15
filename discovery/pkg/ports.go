package pkg

import (
	"fmt"

	"google.golang.org/grpc"
	proto "pulwar.isi.edu/sabres/orchestrator/discovery/protocol"
)

var (
	DefaultDiscoveryPort     = 15010
	DefaultMockDiscoveryPort = 15015
	DefaultScanDiscoveryPort = 15020
)

func ToEndpointAddr(server string, port int) string {
	return fmt.Sprintf("%s:%d", server, port)
}

func WithDiscovery(endpoint string, f func(proto.DiscoveryClient) error) error {

	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to moa service: %v", err)
	}
	client := proto.NewDiscoveryClient(conn)
	defer conn.Close()

	return f(client)
}
