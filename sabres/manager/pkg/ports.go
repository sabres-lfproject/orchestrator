package pkg

import (
	"fmt"

	"google.golang.org/grpc"
	"pulwar.isi.edu/sabres/orchestrator/sabres/manager/protocol"
)

var (
	DefaultManagementPort = 15035
)

func Endpoint(server string, port int) string {
	return fmt.Sprintf("%s:%d", server, port)
}

func WithManagement(endpoint string, f func(protocol.ManagerClient) error) error {

	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to moa service: %v", err)
	}
	client := protocol.NewManagerClient(conn)
	defer conn.Close()

	return f(client)
}
