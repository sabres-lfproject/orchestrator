package pkg

import (
	"fmt"

	"google.golang.org/grpc"
	"pulwar.isi.edu/sabres/orchestrator/sabres/network/protocol"
)

var (
	DefaultNetworkPort = 15025
)

func Endpoint(server string, port int) string {
	return fmt.Sprintf("%s:%d", server, port)
}

func WithNetwork(endpoint string, f func(protocol.NetworkClient) error) error {

	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to moa service: %v", err)
	}
	client := protocol.NewNetworkClient(conn)
	defer conn.Close()

	return f(client)
}
